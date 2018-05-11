package containerio

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	runtime "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
)

const (
	// delimiter used in cri logging format.
	delimiter = ' '
	// eof is end-of-line.
	eol = '\n'
	// timestampFormat is the timestamp format used in cri logging format.
	timestampFormat = time.RFC3339Nano
	// pipeBufSize is the system PIPE_BUF size, on linux it is 4096 bytes.
	pipeBufSize = 4096
	// bufSize is the size of the read buffer.
	bufSize = pipeBufSize - len(timestampFormat) - len(Stdout) - 2 /*2 delimiter*/ - 1 /*eol*/
)

// StreamType is the type of the stream.
type StreamType string

const (
	// Stdin stream type.
	Stdin StreamType = "stdin"
	// Stdout stream type.
	Stdout StreamType = "stdout"
	// Stderr stream type.
	Stderr StreamType = "stderr"
)

func init() {
	Register(func() Backend {
		return &criLogFile{}
	})
}

type criLogFile struct {
	file          *os.File
	outPipeWriter *io.PipeWriter
	outPipeReader *io.PipeReader
	errPipeWriter *io.PipeWriter
	errPipeReader *io.PipeReader
	closed        bool
}

func (c *criLogFile) Name() string {
	return "cri-log-file"
}

func (c *criLogFile) Init(opt *Option) error {
	c.file = opt.criLogFile
	c.outPipeReader, c.outPipeWriter = io.Pipe()
	c.errPipeReader, c.errPipeWriter = io.Pipe()
	go redirectLogs(c.file, c.outPipeReader, Stdout)
	go redirectLogs(c.file, c.errPipeReader, Stderr)
	return nil
}

func redirectLogs(w io.WriteCloser, r io.ReadCloser, stream StreamType) {
	defer r.Close()
	defer w.Close()
	streamBytes := []byte(stream)
	delimiterBytes := []byte{delimiter}
	partialBytes := []byte(runtime.LogTagPartial)
	fullBytes := []byte(runtime.LogTagFull)
	br := bufio.NewReaderSize(r, bufSize)
	for {
		lineBytes, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				logrus.Infof("finish redirecting log file")
			} else {
				logrus.Errorf("failed to redirect log file: %v", err)
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
		// TODO: maybe lock here?
		_, err = w.Write(data)
		if err != nil {
			logrus.Errorf("failed to write %q log to log file: %v", stream, err)
		}
	}
}

func (c *criLogFile) Out() io.Writer {
	return c.outPipeWriter
}

func (c *criLogFile) Err() io.Writer {
	return c.errPipeWriter
}

func (c *criLogFile) In() io.Reader {
	// Log doesn't need stdin.
	return nil
}

func (c *criLogFile) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	c.outPipeWriter.Close()
	c.errPipeWriter.Close()
	return nil
}
