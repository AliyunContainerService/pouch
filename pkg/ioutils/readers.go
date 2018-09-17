package ioutils

import "io"

// ReadCloserWrapper wraps an io.Reader, and implements an io.ReadCloser
// It calls the given callback function when closed. It should be constructed
// with NewReadCloserWrapper
type ReadCloserWrapper struct {
	io.Reader
	closer func() error
}

// NewReadCloserWrapper returns a new io.ReadCloser.
func NewReadCloserWrapper(r io.Reader, closer func() error) io.ReadCloser {
	return &ReadCloserWrapper{
		Reader: r,
		closer: closer,
	}

}

// Close calls back the passed closer function
func (r *ReadCloserWrapper) Close() error {
	return r.closer()
}
