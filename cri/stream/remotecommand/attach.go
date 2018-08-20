package remotecommand

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Attacher knows how to attach a running container in a pod.
type Attacher interface {
	// Attach attaches to the running container in the pod.
	Attach(ctx context.Context, containerID string, streamOpts *Options, streams *Streams) error
}

// ServeAttach handles requests to attach to a container. After creating/receiving the required
// streams, it delegates the actual attaching to attacher.
func ServeAttach(ctx context.Context, w http.ResponseWriter, req *http.Request, attacher Attacher, container string, streamOpts *Options, idleTimeout, streamCreationTimeout time.Duration, supportedProtocols []string) {
	streamCtx, ok := createStreams(w, req, streamOpts, supportedProtocols, idleTimeout, streamCreationTimeout)
	if !ok {
		// Error is handled by createStreams.
		return
	}
	defer streamCtx.conn.Close()

	err := attacher.Attach(ctx, container, streamOpts, &Streams{
		StreamCh:     make(chan struct{}, 1),
		StdinStream:  streamCtx.stdinStream,
		StdoutStream: streamCtx.stdoutStream,
		StderrStream: streamCtx.stderrStream,
	})
	if err != nil {
		err = fmt.Errorf("error attaching to container: %v", err)
		streamCtx.writeStatus(NewInternalError(err))
	} else {
		streamCtx.writeStatus(&StatusError{ErrStatus: Status{
			Status: StatusSuccess,
		}})
	}
}
