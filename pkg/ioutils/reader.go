package ioutils

import "io"

type readCloserWrapper struct {
	io.Reader
	closeFunc func() error
}

func (r *readCloserWrapper) Close() error {
	return r.closeFunc()
}

// NewReadCloserWrapper provides the ability to handle the cleanup during closer.
func NewReadCloserWrapper(r io.Reader, closeFunc func() error) io.ReadCloser {
	return &readCloserWrapper{
		Reader:    r,
		closeFunc: closeFunc,
	}
}
