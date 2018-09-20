package ioutils

import "io"

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
