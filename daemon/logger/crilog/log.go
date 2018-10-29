package crilog

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

const (
	// delimiter used in cri logging format.
	delimiter = ' '
	// eol is end-of-line.
	eol = '\n'
	// timestampFormat is the timestamp format used in cri logging format.
	timestampFormat = time.RFC3339Nano
	// pipeBufSize is the system PIPE_BUF size, on linux it is 4096 bytes.
	pipeBufSize = 4096
	// bufSize is the size of the read buffer.
	bufSize = pipeBufSize - len(timestampFormat) - len(streamStdout) - 2 /*2 delimiter*/ - 1 /*eol*/
	// redirectLogCloseTimeout is used to wait for redirectLogs
	redirectLogCloseTimeout = 10 * time.Second
)

// streamType is the type of the stream.
type streamType string

const (
	// streamStdout stream type.
	streamStdout streamType = "stdout"
	// streamStderr stream type.
	streamStderr streamType = "stderr"
)

// Log represents cri log driver.
//
// NOTE: it might be changed caused by ReopenLog API.
type Log struct {
	Stdout, Stderr io.WriteCloser
	closeFn        func()
}

// Close closes CRI log.
func (l *Log) Close() error {
	if l.Stdout != nil {
		l.Stdout.Close()
	}

	if l.Stderr != nil {
		l.Stderr.Close()
	}

	if l.closeFn != nil {
		l.closeFn()
	}
	return nil
}

// New returns WriteCloser for stream.
func New(path string, withTerminal bool) (*Log, error) {
	// TODO(fuweid): need to serialize writer since both the stdout and
	// stderr share the same writer.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return nil, err
	}

	var (
		stdoutw, stderrw           io.WriteCloser
		stdoutStopCh, stderrStopCh <-chan struct{}
	)

	stdoutw, stdoutStopCh = newCRILogger(path, f, streamStdout)
	if !withTerminal {
		stderrw, stderrStopCh = newCRILogger(path, f, streamStderr)
	}

	closeFn := func() {
		if stdoutStopCh != nil {
			select {
			case <-stdoutStopCh:
			case <-time.After(redirectLogCloseTimeout):
				logrus.WithField("cri-log", path).
					Warn("failed to stop stdout's redirectLogs")
			}
		}

		if stderrStopCh != nil {
			select {
			case <-stderrStopCh:
			case <-time.After(redirectLogCloseTimeout):
				logrus.WithField("cri-log", path).
					Warn("failed to stop stderr's redirectLogs")
			}
		}
		f.Close()
	}

	return &Log{
		Stdout:  stdoutw,
		Stderr:  stderrw,
		closeFn: closeFn,
	}, nil
}

func newCRILogger(path string, w io.Writer, typ streamType) (io.WriteCloser, <-chan struct{}) {
	stopCh := make(chan struct{})
	pir, piw := io.Pipe()
	go func() {
		redirectLogs(path, w, pir, typ)
		close(stopCh)
	}()
	return piw, stopCh
}

func redirectLogs(path string, w io.Writer, r io.ReadCloser, stream streamType) {
	defer r.Close()

	streamBytes := []byte(stream)
	delimiterBytes := []byte{delimiter}
	partialBytes := []byte(runtime.LogTagPartial)
	fullBytes := []byte(runtime.LogTagFull)
	br := bufio.NewReaderSize(r, bufSize)
	for {
		lineBytes, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				logrus.Infof("finish redirecting log file(name=%v)", path)
			} else {
				logrus.WithError(err).Errorf("failed to redirect log file(name=%v)", path)
			}
			return
		}
		tagBytes := fullBytes
		if isPrefix {
			tagBytes = partialBytes
		}
		timestampBytes := time.Now().AppendFormat(nil, time.RFC3339Nano)
		data := bytes.Join([][]byte{timestampBytes, streamBytes, tagBytes, lineBytes}, delimiterBytes)
		data = append(data, eol)

		if _, err := w.Write(data); err != nil {
			logrus.Errorf("failed to write %q log to log file: %v", stream, err)
		}
	}
}
