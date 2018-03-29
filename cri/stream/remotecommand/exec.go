package remotecommand

import (
	"fmt"
	"net/http"
	"time"
)

// Executor knows how to execute a command in a container of the pod.
type Executor interface {
	// Exec executes a command in a container of the pod.
	Exec(containerID string, cmd []string, streamOpts *Options, streams *Streams) (uint32, error)
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

	exitCode, err := executor.Exec(container, cmd, streamOpts, &Streams{
		StdinStream:  ctx.stdinStream,
		StdoutStream: ctx.stdoutStream,
		StderrStream: ctx.stderrStream,
	})
	if err != nil {
		err = fmt.Errorf("error executing command in container: %v", err)
		ctx.writeStatus(NewInternalError(err))
	} else if exitCode != 0 {
		ctx.writeStatus(&StatusError{ErrStatus: Status{
			Status: StatusFailure,
			Reason: NonZeroExitCodeReason,
			Details: &StatusDetails{
				Causes: []StatusCause{
					{
						Type:    ExitCodeCauseType,
						Message: fmt.Sprintf("%d", exitCode),
					},
				},
			},
			Message: fmt.Sprintf("command terminated with non-zero exit code"),
		}})
	} else {
		ctx.writeStatus(&StatusError{ErrStatus: Status{
			Status: StatusSuccess,
		}})
	}
}
