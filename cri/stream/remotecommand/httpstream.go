package remotecommand

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	apitypes "github.com/alibaba/pouch/apis/types"
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
type streamContext struct {
	conn         io.Closer
	stdinStream  io.ReadCloser
	stdoutStream io.WriteCloser
	stderrStream io.WriteCloser
	resizeStream io.ReadCloser
	resizeChan   chan apitypes.ResizeOptions
	writeStatus  func(status *StatusError) error
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

func createStreams(w http.ResponseWriter, req *http.Request, opts *Options, supportedStreamProtocols []string, idleTimeout time.Duration, streamCreationTimeout time.Duration) (*streamContext, bool) {
	var ctx *streamContext
	var ok bool
	if wsstream.IsWebSocketRequest(req) {
		ctx, ok = createWebSocketStreams(w, req, opts, idleTimeout)
	} else {
		ctx, ok = createHTTPStreamStreams(w, req, opts, supportedStreamProtocols, idleTimeout, streamCreationTimeout)
	}
	if !ok {
		return nil, false
	}

	if ctx.resizeStream != nil {
		ctx.resizeChan = make(chan apitypes.ResizeOptions)
		go handleResizeEvents(ctx.resizeStream, ctx.resizeChan)
	}

	return ctx, true
}

func handleResizeEvents(stream io.Reader, channel chan<- apitypes.ResizeOptions) {
	var err error
	decoder := json.NewDecoder(stream)
	for {
		size := apitypes.ResizeOptions{}
		err = decoder.Decode(&size)
		if err != nil {
			break
		}
		channel <- size
	}
	if err != io.EOF {
		logrus.Errorf("failed to decode resize request from resize stream: %v", err)
	}
	close(channel)
}

func createHTTPStreamStreams(w http.ResponseWriter, req *http.Request, opts *Options, supportedStreamProtocols []string, idleTimeout time.Duration, streamCreationTimeout time.Duration) (*streamContext, bool) {
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
	case constant.StreamProtocolV1Name:
		handler = &v1ProtocolHandler{}
	case constant.StreamProtocolV2Name:
		handler = &v2ProtocolHandler{}
	case constant.StreamProtocolV3Name:
		handler = &v3ProtocolHandler{}
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
	if opts.TTY && handler.supportsTerminalResizing() {
		expectedStreams++
	}

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
	waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streamContext, error)
	// supportsTerminalResizing returns true if the protocol handler supports terminal resizing.
	supportsTerminalResizing() bool
}

// v3ProtocolHandler implements the V3 protocol version for streaming command execution.
type v3ProtocolHandler struct{}

func (*v3ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streamContext, error) {
	ctx := &streamContext{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
done:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(constant.StreamType)
			switch streamType {
			case constant.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)
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
			case constant.StreamTypeResize:
				ctx.resizeStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				logrus.Errorf("Unexpected stream type: %q", streamType)
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break done
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, fmt.Errorf("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

// supportsTerminalResizing returns true because v3ProtocolHandler supports it.
func (*v3ProtocolHandler) supportsTerminalResizing() bool { return true }

// v2ProtocolHandler implements the V2 protocol version for streaming command execution.
type v2ProtocolHandler struct{}

func (*v2ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streamContext, error) {
	ctx := &streamContext{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
done:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(constant.StreamType)
			switch streamType {
			case constant.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)
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
				break done
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, fmt.Errorf("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

// supportsTerminalResizing returns false because v2ProtocolHandler doesn't support it.
func (*v2ProtocolHandler) supportsTerminalResizing() bool { return false }

// v1ProtocolHandler implements the V1 protocol version for streaming command execution.
type v1ProtocolHandler struct{}

func (*v1ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streamContext, error) {
	ctx := &streamContext{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
done:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(constant.StreamType)
			switch streamType {
			case constant.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)
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
				break done
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

// supportsTerminalResizing returns false because v1ProtocolHandler doesn't support it.
func (*v1ProtocolHandler) supportsTerminalResizing() bool { return false }

func v1WriteStatusFunc(stream io.Writer) func(status *StatusError) error {
	return func(status *StatusError) error {
		if status.Status().Status == StatusSuccess {
			return nil
		}

		_, err := stream.Write([]byte(status.Error()))
		return err
	}
}

func v4WriteStatusFunc(stream io.Writer) func(status *StatusError) error {
	return func(status *StatusError) error {
		bs, err := json.Marshal(status.Status())
		if err != nil {
			return err
		}

		_, err = stream.Write(bs)
		return err
	}
}
