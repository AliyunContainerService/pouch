package portforward

import (
	"net/http"
	"time"
	"sync"
	"strconv"
	"fmt"

	"github.com/alibaba/pouch/cri/stream/constant"
	"github.com/alibaba/pouch/cri/stream/httpstream"
	"github.com/alibaba/pouch/cri/stream/httpstream/spdy"

	"github.com/sirupsen/logrus"
)

// httpStreamReceived is the httpstream.NewStreamHandler for port
// forward streams. It checks each stream's port and stream type headers,
// rejecting any streams that with missing or invalid values. Each valid
// stream is sent to the streams channel.
func httpStreamReceived(streams chan httpstream.Stream) func(httpstream.Stream, <-chan struct{}) error {
	return func(stream httpstream.Stream, replySent <-chan struct{}) error {
		// Make sure it has a valid port header.
		portString := stream.Headers().Get(constant.PortHeader)
		if len(portString) == 0 {
			return fmt.Errorf("%q header is required", constant.PortHeader)
		}
		port, err := strconv.ParseUint(portString, 10, 16)
		if err != nil {
			return fmt.Errorf("unable to parse %q as a port: %v", portString, err)
		}
		if port < 1 {
			return fmt.Errorf("port %q must be > 0", portString)
		}

		// Make sure it has a valid stream type header.
		streamType := stream.Headers().Get(constant.StreamType)
		if len(streamType) == 0 {
			return fmt.Errorf("%q header is required", constant.StreamType)
		}
		if streamType != constant.StreamTypeError && streamType != constant.StreamTypeData {
			return fmt.Errorf("invalid stream type %q", streamType)
		}

		streams <- stream
		return nil
	}
}

func handleHttpStreams(w http.ResponseWriter, req *http.Request, portForwarder PortForwarder, podName string, idleTimeout, streamCreationTimeout time.Duration, supportedPortForwardProtocols []string) error {
	_, err := httpstream.Handshake(w, req, supportedPortForwardProtocols)
	// Negotiated protocol isn't currently used server side, but could be in the future.
	if err != nil {
		// Handshake writes the error to the client
		return err
	}
	streamChan := make(chan httpstream.Stream, 1)

	logrus.Infof("Upgrading port forward response")
	upgrader := spdy.NewResponseUpgrader()
	conn := upgrader.UpgradeResponse(w, req, httpStreamReceived(streamChan))
	if conn == nil {
		return fmt.Errorf("unable to upgrade connection")
	}
	defer conn.Close()

	logrus.Infof("(conn=%p) setting forwarding streaming connection idle timeout to %v", conn, idleTimeout)
	conn.SetIdleTimeout(idleTimeout)

	h := &httpStreamHandler{
		conn:					conn,
		streamChan:				streamChan,
		streamPairs:			make(map[string]*httpStreamPair),
		streamCreationTimeout:	streamCreationTimeout,
		pod:					podName,
		forwarder:				portForwarder,
	}
	h.run()

	return nil
}

// httpStreamHandler is capable of processing multiple port forward
// requests over a single httpstream.Connection.
type httpStreamHandler struct {
	conn 					httpstream.Connection
	streamChan				chan httpstream.Stream
	streamPairsLock			sync.RWMutex
	streamPairs 			map[string]*httpStreamPair
	streamCreationTimeout	time.Duration
	pod 					string
	forwarder 				PortForwarder
}

// getStreamPair returns a httpStreamPair for requestID. This creates a
// new pair if one does not yet exist for the requestID. The returned bool is
// true if the pair was created.
func (h *httpStreamHandler) getStreamPair(requestID string) (*httpStreamPair, bool) {
	h.streamPairsLock.Lock()
	defer h.streamPairsLock.Unlock()

	if p, ok := h.streamPairs[requestID]; ok {
		logrus.Infof("(conn=%p, request=%s) found existing stream pair", h.conn, requestID)
		return p, false
	}

	logrus.Infof("(conn=%p, request=%s) creating new stream pair", h.conn, requestID)

	p := newPortForwardPair(requestID)
	h.streamPairs[requestID] = p

	return p, true
}

// hasStreamPair returns a bool indicating if a stream pair for requestID exists.
func (h *httpStreamHandler) hasStreamPair(requestID string) bool {
	h.streamPairsLock.RLock()
	defer h.streamPairsLock.RUnlock()

	_, ok := h.streamPairs[requestID]

	return ok
}

// removeStreamPair removes the stream pair identified by requestID from streamPairs.
func (h *httpStreamHandler) removeStreamPair(requestID string) {
	h.streamPairsLock.Lock()
	defer h.streamPairsLock.Unlock()

	delete(h.streamPairs, requestID)
}

// monitorStreamPair waits for the pair to receive both its error and data
// streams, or for the timeout to expire (whichever happens first), and then
// removes the pair.
func (h *httpStreamHandler) monitorStreamPair(p *httpStreamPair, timeout <-chan time.Time) {
	select {
	case <-timeout:
		msg := fmt.Sprintf("(conn=%v, request=%s) timed out waiting for streams", h.conn, p.requestID)
		p.printError(msg)
	case <-p.complete:
		logrus.Infof("(conn=%v, request=%s) successfully received error and data streams", h.conn, p.requestID)
	}
	h.removeStreamPair(p.requestID)
}

// requestID returns the request id for stream.
func (h *httpStreamHandler) requestID(stream httpstream.Stream) string {
	requestID := stream.Headers().Get(constant.PortForwardRequestIDHeader)
	if len(requestID) == 0 {
		// TODO: support the connection come from the older client
		// that isn't generating the request id header.
	}

	return requestID
}

// run is the main loop for the httpStreamHandler. It process new streams,
// invoking portForward for each complete stream pair. The loop exits
// when the httpstream.Connection is closed.
func (h *httpStreamHandler) run() {
	logrus.Infof("(conn=%p) waiting for port forward streams", h.conn)
Loop:
	for {
		select {
		case <-h.conn.CloseChan():
			logrus.Infof("(conn=%p) upgraded connection closed", h.conn)
			break Loop
		case stream := <-h.streamChan:
			requestID := h.requestID(stream)
			streamType := stream.Headers().Get(constant.StreamType)
			logrus.Infof("(conn=%p, request=%s) received new stream of type %s", h.conn, requestID, streamType)

			p, created := h.getStreamPair(requestID)
			if created {
				go h.monitorStreamPair(p, time.After(h.streamCreationTimeout))
			}
			if complete, err := p.add(stream); err != nil {
				msg := fmt.Sprintf("error processing stream for request %s: %v", requestID, err)
				p.printError(msg)
			} else if complete {
				go h.portForward(p)
			}
		}
	}
}

// portForward invokes the httpStreamHandler's forwarder.PortForward
// function for the given stream pair.
func (h *httpStreamHandler) portForward(p *httpStreamPair) {
		defer p.dataStream.Close()
		defer p.errorStream.Close()

		portString := p.dataStream.Headers().Get(constant.PortHeader)
		port, _ := strconv.ParseInt(portString, 10, 32)

		logrus.Infof("(conn=%p, request=%s) invoking forwarder.PortForward for port %s", h.conn, p.requestID, portString)
		err := h.forwarder.PortForward(h.pod, int32(port), p.dataStream)
		logrus.Infof("(conn=%p, request=%s) done invoking forwarder.PortForward for port %s", h.conn, p.requestID, portString)

		if err != nil {
			msg := fmt.Sprintf("error forwarding port %d to pod %s: %v", port, h.pod, err)
			p.printError(msg)
		}
}

// httpStreamPair represents the error and data streams for a port
// forwarding request.
type httpStreamPair struct {
	lock 		sync.RWMutex
	requestID	string
	dataStream	httpstream.Stream
	errorStream	httpstream.Stream
	complete	chan struct{}
}

// newPortForwardPair creates a new httpStreamPair.
func newPortForwardPair(requestID string) *httpStreamPair {
	return &httpStreamPair{
		requestID:	requestID,
		complete:	make(chan struct{}),
	}
}

// add adds the stream to the httpStreamPair. If the pair already
// contains a stream for the new stream's type, an error is returned. add
// returns true if both the data and error streams for this pair have been
// received.
func (p *httpStreamPair) add(stream httpstream.Stream) (bool, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	switch stream.Headers().Get(constant.StreamType) {
	case constant.StreamTypeError:
		if p.errorStream != nil {
			return false, fmt.Errorf("error stream already assigned")
		}
		p.errorStream = stream
	case constant.StreamTypeData:
		if p.dataStream != nil {
			return false, fmt.Errorf("data stream already assigned")
		}
		p.dataStream = stream
	}

	complete := p.errorStream != nil && p.dataStream != nil
	if complete {
		close(p.complete)
	}

	return complete, nil
}

// printError writes s to p.errorStream if p.errorStream has been set.
func (p *httpStreamPair) printError(s string) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if p.errorStream != nil {
		fmt.Fprint(p.errorStream, s)
	}
}
