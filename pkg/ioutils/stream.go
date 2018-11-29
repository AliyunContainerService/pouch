package ioutils

import (
	"errors"
	"io"
	"os"

	"github.com/docker/docker/pkg/term"
	"github.com/sirupsen/logrus"
)

// Streams is an interface which exposes the standard input and output streams
type Streams interface {
	In() *InStream
	Out() *OutStream
	Err() io.Writer
}

// CliStream represents stream wrapper for cli
type CliStream struct {
	InStream  *InStream
	OutStream *OutStream
	ErrStream io.Writer
}

// In returns InStream
func (c *CliStream) In() *InStream {
	return c.InStream
}

// Out returns OutStream
func (c *CliStream) Out() *OutStream {
	return c.OutStream
}

// Err returns ErrStream
func (c *CliStream) Err() io.Writer {
	return c.ErrStream
}

// CommonStream is a stream utils
type CommonStream struct {
	fd         uintptr
	isTerminal bool
	state      *term.State
}

// FD returns the file descriptor number for this stream
func (s *CommonStream) FD() uintptr {
	return s.fd
}

// IsTerminal returns true if this stream is connected to a terminal
func (s *CommonStream) IsTerminal() bool {
	return s.isTerminal
}

// RestoreTerminal restores normal mode to the terminal
func (s *CommonStream) RestoreTerminal() {
	if s.state != nil {
		term.RestoreTerminal(s.fd, s.state)
	}
}

// SetIsTerminal sets the boolean used for isTerminal
func (s *CommonStream) SetIsTerminal(isTerminal bool) {
	s.isTerminal = isTerminal
}

// InStream is an input stream used to read user input
type InStream struct {
	CommonStream
	in io.ReadCloser
}

func (i *InStream) Read(p []byte) (int, error) {
	return i.in.Read(p)
}

// Close closes io in
func (i *InStream) Close() error {
	return i.in.Close()
}

// SetRawTerminal sets raw mode on the input terminal
func (i *InStream) SetRawTerminal() (err error) {
	if os.Getenv("NORAW") != "" || !i.CommonStream.isTerminal {
		return nil
	}
	i.CommonStream.state, err = term.SetRawTerminal(i.CommonStream.fd)
	return err
}

// CheckTty checks if we are trying to attach to a container tty
func (i *InStream) CheckTty(attachStdin, ttyMode bool) error {
	if ttyMode && attachStdin && !i.isTerminal {
		return errors.New("the input device is not a TTY")
	}
	return nil
}

// NewInStream returns a new InStream object from a ReadCloser
func NewInStream(in io.ReadCloser) *InStream {
	fd, isTerminal := term.GetFdInfo(in)
	return &InStream{CommonStream: CommonStream{fd: fd, isTerminal: isTerminal}, in: in}
}

// OutStream is an output stream used to write output.
type OutStream struct {
	CommonStream
	out io.Writer
}

func (o *OutStream) Write(p []byte) (int, error) {
	return o.out.Write(p)
}

// SetRawTerminal sets raw mode on the input terminal
func (o *OutStream) SetRawTerminal() (err error) {
	if os.Getenv("NORAW") != "" || !o.CommonStream.isTerminal {
		return nil
	}
	o.CommonStream.state, err = term.SetRawTerminalOutput(o.CommonStream.fd)
	return err
}

// GetTtySize returns the height and width in characters of the tty
func (o *OutStream) GetTtySize() (uint, uint) {
	if !o.isTerminal {
		return 0, 0
	}
	ws, err := term.GetWinsize(o.fd)
	if err != nil {
		logrus.Debugf("Error getting size: %s", err)
		if ws == nil {
			return 0, 0
		}
	}
	return uint(ws.Height), uint(ws.Width)
}

// NewOutStream returns a new OutStream object from a Writer
func NewOutStream(out io.Writer) *OutStream {
	fd, isTerminal := term.GetFdInfo(out)
	return &OutStream{CommonStream: CommonStream{fd: fd, isTerminal: isTerminal}, out: out}
}
