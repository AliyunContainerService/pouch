package remotecommand

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alibaba/pouch/cri/stream/constant"
	"github.com/alibaba/pouch/cri/stream/httpstream"
	"github.com/alibaba/pouch/cri/stream/httpstream/spdy"

	"github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/util/wsstream"
)

// Options contains details about which streams are required for
// remote command execution.
type Options struct {
	Stdin  bool
	Stdout bool
	Stderr bool
	TTY    bool
}

// Streams contains all the streams used to stdio for
// remote command execution.
type Streams struct {
	// Notified from StreamCh if streams broken.
	StreamCh     chan struct{}
	StdinStream  io.ReadCloser
	StdoutStream io.WriteCloser
	StderrStream io.WriteCloser
}

// context contains the connection and streams used when
// forwarding an attach or execute session into a container.
type context struct {
	conn         io.Closer
	stdinStream  io.ReadCloser
	stdoutStream io.WriteCloser
	stderrStream io.WriteCloser
	resizeStream io.ReadCloser
	tty          bool
}

// streamAndReply holds both a Stream and a channel that is closed when the stream's reply frame is
// enqueued. Consumers can wait for replySent to be closed prior to proceeding, to ensure that the
// replyFrame is enqueued before the connection's goaway frame is sent (e.g. if a stream was
// received and right after, the connection gets closed).
type streamAndReply struct {
	httpstream.Stream
	replySent <-chan struct{}
}

func createStreams(w http.ResponseWriter, req *http.Request, opts *Options, supportedStreamProtocols []string, idleTimeout time.Duration, streamCreationTimeout time.Duration) (*context, bool) {
	var ctx *context
	var ok bool
	if wsstream.IsWebSocketRequest(req) {
		ctx, ok = createWebSocketStreams(w, req, opts, idleTimeout)
	} else {
		ctx, ok = createHTTPStreamStreams(w, req, opts, supportedStreamProtocols, idleTimeout, streamCreationTimeout)
	}
	if !ok {
		return nil, false
	}

	// TODO: resizeStream support.

	return ctx, true
}

func createHTTPStreamStreams(w http.ResponseWriter, req *http.Request, opts *Options, supportedStreamProtocols []string, idleTimeout time.Duration, streamCreationTimeout time.Duration) (*context, bool) {
	protocol, err := httpstream.Handshake(w, req, supportedStreamProtocols)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return nil, false
	}

	streamCh := make(chan streamAndReply)

	upgrader := spdy.NewResponseUpgrader()
	conn := upgrader.UpgradeResponse(w, req, func(stream httpstream.Stream, replySent <-chan struct{}) error {
		streamCh <- streamAndReply{Stream: stream, replySent: replySent}
		return nil
	})
	// From this point on, we can no longer call methods on response.
	if conn == nil {
		// The upgrader is responsible for notifying the client of any errors that
		// occurred during upgrading. All we can do is return here at this point
		// if we weren't successful in upgrading.
		return nil, false
	}

	conn.SetIdleTimeout(idleTimeout)

	var handler protocolHandler
	switch protocol {
	case "":
		logrus.Infof("Client did not request protocol negotiation. Falling back to %q", constant.StreamProtocolV1Name)
		fallthrough
	case constant.StreamProtocolV2Name:
		handler = &v2ProtocolHandler{}
	case constant.StreamProtocolV1Name:
		handler = &v1ProtocolHandler{}
	}

	// Count the streams client asked for, starting with 1.
	expectedStreams := 1
	if opts.Stdin {
		expectedStreams++
	}
	if opts.Stdout {
		expectedStreams++
	}
	if opts.Stderr {
		expectedStreams++
	}
	// TODO: handle opts.TTY

	expired := time.NewTimer(streamCreationTimeout)
	defer expired.Stop()

	ctx, err := handler.waitForStreams(streamCh, expectedStreams, expired.C)
	if err != nil {
		logrus.Errorf("failed to create streams: %v", err)
		return nil, false
	}

	ctx.conn = conn

	return ctx, true
}

// waitStreamReply waits until either replySent or stop is closed. If replySent is closed, it sends
// an empty struct to the notify channel.
func waitStreamReply(replySent <-chan struct{}, notify chan<- struct{}, stop <-chan struct{}) {
	select {
	case <-replySent:
		notify <- struct{}{}
	case <-stop:
	}
}

type protocolHandler interface {
	// waitForStreams waits for the expected streams or a timeout, returning a
	// remoteCommandContext if all the streams were received, or an error if not.
	waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*context, error)
}

// v2ProtocolHandler implements the V2 protocol version for streaming command execution.
type v2ProtocolHandler struct{}

func (*v2ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*context, error) {
	ctx := &context{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(constant.StreamType)
			switch streamType {
			case constant.StreamTypeError:
				go waitStreamReply(stream.replySent, replyChan, stop)
			case constant.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case constant.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case constant.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				logrus.Errorf("Unexpected stream type: %q", streamType)
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, fmt.Errorf("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

// v1ProtocolHandler implements the V1 protocol version for streaming command execution.
type v1ProtocolHandler struct{}

func (*v1ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*context, error) {
	ctx := &context{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(constant.StreamType)
			switch streamType {
			case constant.StreamTypeError:
				// ctx.writeStatus = v1WriteStatusFunc(stream)
				go waitStreamReply(stream.replySent, replyChan, stop)
			case constant.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case constant.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case constant.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				logrus.Errorf("Unexpected stream type: %q", streamType)
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, fmt.Errorf("timed out waiting for client to create streams")
		}
	}

	if ctx.stdinStream != nil {
		ctx.stdinStream.Close()
	}

	return ctx, nil
}
