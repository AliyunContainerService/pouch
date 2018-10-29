package ioutils

import "io"

// noopWriter is an io.Writer on which all Write calls succeed without
// doing anything.
type noopWriter struct{}

func (nw *noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (nw *noopWriter) Close() error {
	return nil
}

// NewNoopWriteCloser returns the no-op WriteCloser.
func NewNoopWriteCloser() io.WriteCloser {
	return &noopWriter{}
}

type writeCloserWrapper struct {
	io.Writer
	closeFunc func() error
}

func (w *writeCloserWrapper) Close() error {
	return w.closeFunc()
}

// NewWriteCloserWrapper provides the ability to handle the cleanup during closer.
func NewWriteCloserWrapper(w io.Writer, closeFunc func() error) io.WriteCloser {
	return &writeCloserWrapper{
		Writer:    w,
		closeFunc: closeFunc,
	}
}

// CloseWriter is an interface which represents the implementation closes the
// writing side of writer.
type CloseWriter interface {
	CloseWrite() error
}
