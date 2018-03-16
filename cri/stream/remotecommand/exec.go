package remotecommand

import (
	"net/http"
	"time"
)

// Executor knows how to execute a command in a container of the pod.
type Executor interface {
	// Exec executes a command in a container of the pod.
	Exec(containerID string, cmd []string, streamOpts *Options, streams *Streams) error
}

// ServeExec handles requests to execute a command in a container. After
// creating/receiving the required streams, it delegates the actual execution
// to the executor.
func ServeExec(w http.ResponseWriter, req *http.Request, executor Executor, container string, cmd []string, streamOpts *Options, supportedProtocols []string, idleTimeout time.Duration, streamCreationTimeout time.Duration) {
	ctx, ok := createStreams(w, req, streamOpts, supportedProtocols, idleTimeout, streamCreationTimeout)
	if !ok {
		// Error is handled by createStreams.
		return
	}
	defer ctx.conn.Close()

	executor.Exec(container, cmd, streamOpts, &Streams{
		StdinStream:  ctx.stdinStream,
		StdoutStream: ctx.stdoutStream,
		StderrStream: ctx.stderrStream,
	})
}
