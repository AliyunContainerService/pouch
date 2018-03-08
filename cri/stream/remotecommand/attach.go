package remotecommand

import (
	"net/http"
	"time"
)

// Attacher knows how to attach a running container in a pod.
type Attacher interface {
	// Attach attaches to the running container in the pod.
	Attach(containerID string, streamOpts *Options, streams *Streams) error
}

// ServeAttach handles requests to attach to a container. After creating/receiving the required
// streams, it delegates the actual attaching to attacher.
func ServeAttach(w http.ResponseWriter, req *http.Request, attacher Attacher, container string, streamOpts *Options, idleTimeout, streamCreationTimeout time.Duration, supportedProtocols []string) {
	ctx, ok := createStreams(w, req, streamOpts, supportedProtocols, idleTimeout, streamCreationTimeout)
	if !ok {
		// Error is handled by createStreams.
		return
	}
	defer ctx.conn.Close()

	attacher.Attach(container, streamOpts, &Streams{
		StreamCh:     make(chan struct{}, 1),
		StdinStream:  ctx.stdinStream,
		StdoutStream: ctx.stdoutStream,
		StderrStream: ctx.stderrStream,
	})
}
