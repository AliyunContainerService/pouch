package server

import (
	"bytes"
	"io"
	"os"
)

type filterLogWriter struct {
	io.Writer
}

var stdFilterLogWriter filterLogWriter
var filterString = []byte("http: TLS handshake error from")

func (w filterLogWriter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, filterString) {
		return 0, nil
	}
	return os.Stderr.Write(p)
}
