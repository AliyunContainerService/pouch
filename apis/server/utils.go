package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sirupsen/logrus"
)

type flusher interface {
	// Flush sends any buffered data to the client.
	Flush()
}

// writeFlusher will flush data after every successful write.
type writeFlusher struct {
	w io.Writer
	f flusher
}

// Write writes and flushs data.
func (wf *writeFlusher) Write(p []byte) (int, error) {
	n, err := wf.w.Write(p)
	if err != nil {
		return n, err
	}
	wf.f.Flush()
	return n, nil
}

// newWriteFlusher will new io.Writer which flushs data after successful write
// if w has Flush() implementation.
func newWriteFlusher(w io.Writer) io.Writer {
	if f, ok := w.(flusher); ok {
		return &writeFlusher{w: w, f: f}
	}
	return w
}

// writeLogStream will convert to WriteFlusher to writer log.
func writeLogStream(ctx context.Context, w io.Writer, tty bool, opt *types.ContainerLogsOptions, msgs <-chan *logger.LogMessage) {
	// NOTE: The default HTTP/1.x and HTTP/2 ResponseWriter implementations Flusher.
	wf := newWriteFlusher(w)

	stdoutStream, stderrStream := wf, wf
	// NOTE: compatible with docker API
	if !tty {
		stdoutStream = stdcopy.NewStdWriter(stdoutStream, stdcopy.Stdout)
		stderrStream = stdcopy.NewStdWriter(stderrStream, stdcopy.Stderr)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}

			// check if the message contains an error.
			// if so, write that error and exit
			if msg.Err != nil {
				fmt.Fprintf(stderrStream, "unexpected error during reading logs: %v\n", msg.Err)
				return
			}

			logLine := msg.Line

			if opt.Details && len(msg.Attrs) > 0 {
				var ss []string
				for k, v := range msg.Attrs {
					ss = append(ss, k+"="+v)
				}
				// keep the log attrs sorted
				sort.Slice(ss, func(i, j int) bool {
					keyI := strings.Split(ss[i], "=")
					keyJ := strings.Split(ss[j], "=")
					return keyI[0] < keyJ[0]
				})
				logLine = append([]byte(strings.Join(ss, ",")+" "), logLine...)
			}

			if opt.Timestamps {
				logLine = append([]byte(msg.Timestamp.Format(utils.TimeLayout)+" "), logLine...)
			}

			if msg.Source == "stdout" && opt.ShowStdout {
				if _, err := stdoutStream.Write(logLine); err != nil {
					logrus.Errorf("unexpected error during stdout log: %v\n", err)
					return
				}
			}
			if msg.Source == "stderr" && opt.ShowStderr {
				if _, err := stderrStream.Write(logLine); err != nil {
					logrus.Errorf("unexpected error during stderr log: %v\n", err)
				}
			}
		}
	}
}

// logCreateOptions will print create args in pouchd logs for debugging.
func logCreateOptions(objType string, config interface{}) {
	args, err := json.Marshal(config)
	if err != nil {
		logrus.Errorf("failed to marsal config for %s: %v", objType, err)
	} else {
		logrus.Infof("create %s with args: %v", objType, string(args))
	}
}
