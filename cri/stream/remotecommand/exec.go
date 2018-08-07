package remotecommand

import (
	"context"
	"fmt"
	"net/http"
	"time"

	apitypes "github.com/alibaba/pouch/apis/types"
)

// Executor knows how to execute a command in a container of the pod.
type Executor interface {
	// Exec executes a command in a container of the pod.
	Exec(ctx context.Context, containerID string, cmd []string, resizeChan <-chan apitypes.ResizeOptions, streamOpts *Options, streams *Streams) (uint32, error)
}

// ServeExec handles requests to execute a command in a container. After
// creating/receiving the required streams, it delegates the actual execution
// to the executor.
func ServeExec(ctx context.Context, w http.ResponseWriter, req *http.Request, executor Executor, container string, cmd []string, streamOpts *Options, supportedProtocols []string, idleTimeout time.Duration, streamCreationTimeout time.Duration) {
	streamCtx, ok := createStreams(w, req, streamOpts, supportedProtocols, idleTimeout, streamCreationTimeout)
	if !ok {
		// Error is handled by createStreams.
		return
	}
	defer streamCtx.conn.Close()

	exitCode, err := executor.Exec(ctx, container, cmd, streamCtx.resizeChan, streamOpts, &Streams{
		StdinStream:  streamCtx.stdinStream,
		StdoutStream: streamCtx.stdoutStream,
		StderrStream: streamCtx.stderrStream,
	})
	if err != nil {
		err = fmt.Errorf("error executing command in container: %v", err)
		streamCtx.writeStatus(NewInternalError(err))
	} else if exitCode != 0 {
		streamCtx.writeStatus(&StatusError{ErrStatus: Status{
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
		streamCtx.writeStatus(&StatusError{ErrStatus: Status{
			Status: StatusSuccess,
		}})
	}
}
