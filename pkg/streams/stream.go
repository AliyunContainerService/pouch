package streams

import (
	"io"
	"sync"

	"github.com/alibaba/pouch/pkg/ioutils"
	"github.com/alibaba/pouch/pkg/multierror"
)

// NewStream returns new streams.
func NewStream() *Stream {
	return &Stream{
		stdout: &multiWriter{},
		stderr: &multiWriter{},
	}
}

// Stream is used to handle container IO.
type Stream struct {
	sync.WaitGroup
	stdin          io.ReadCloser
	stdinPipe      io.WriteCloser
	stdout, stderr *multiWriter
}

// Stdin returns the Stdin for reader.
func (s *Stream) Stdin() io.ReadCloser {
	return s.stdin
}

// StdinPipe returns the Stdin for writer.
func (s *Stream) StdinPipe() io.WriteCloser {
	return s.stdinPipe
}

// NewStdinInput creates pipe for Stdin() and StdinPipe().
func (s *Stream) NewStdinInput() {
	s.stdin, s.stdinPipe = io.Pipe()
}

// NewDiscardStdinInput creates a no-op WriteCloser for StdinPipe().
func (s *Stream) NewDiscardStdinInput() {
	s.stdin, s.stdinPipe = nil, ioutils.NewNoopWriteCloser()
}

// Stdout returns the Stdout for writer.
func (s *Stream) Stdout() io.WriteCloser {
	return s.stdout
}

// Stderr returns the Stderr for writer.
func (s *Stream) Stderr() io.WriteCloser {
	return s.stderr
}

// AddStdoutWriter adds the stdout writer.
func (s *Stream) AddStdoutWriter(w io.WriteCloser) {
	s.stdout.Add(w)
}

// AddStderrWriter adds the stderr writer.
func (s *Stream) AddStderrWriter(w io.WriteCloser) {
	s.stderr.Add(w)
}

// NewStdoutPipe creates pipe and register it into Stdout.
func (s *Stream) NewStdoutPipe() io.ReadCloser {
	r, w := io.Pipe()
	s.stdout.Add(w)
	return r
}

// NewStderrPipe creates pipe and register it into Stderr.
func (s *Stream) NewStderrPipe() io.ReadCloser {
	r, w := io.Pipe()
	s.stderr.Add(w)
	return r
}

// Close closes streams.
func (s *Stream) Close() error {
	multiErrs := new(multierror.Multierrors)

	if s.stdin != nil {
		if err := s.stdin.Close(); err != nil {
			multiErrs.Append(err)
		}
	}

	if err := s.stdout.Close(); err != nil {
		multiErrs.Append(err)
	}

	if err := s.stderr.Close(); err != nil {
		multiErrs.Append(err)
	}

	if multiErrs.Size() > 0 {
		return multiErrs
	}
	return nil
}
