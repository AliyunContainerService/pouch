package containerio

import (
	"io"
)

// writeCloserWrapper combines a writer and a closer to construct a WriteCloser
type writeCloserWrapper struct {
	w io.Writer
	c io.Closer
}

// NewWriteCloserWrapper creates the writeCloserWrapper from a writer and a closer.
func NewWriteCloserWrapper(w io.Writer, c io.Closer) io.WriteCloser {
	return &writeCloserWrapper{
		w: w,
		c: c,
	}
}

// Write passes through the data into the internal writer.
func (w *writeCloserWrapper) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

// Close calls the internal closer.
func (w *writeCloserWrapper) Close() error {
	return w.c.Close()
}
