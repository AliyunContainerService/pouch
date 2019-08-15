package streams

import (
	"io"
	"sync"

	"github.com/alibaba/pouch/pkg/log"
)

// multiWriter allows caller to broadcast data to several writers.
type multiWriter struct {
	sync.Mutex
	writers []io.WriteCloser
}

// Add registers one writer into MultiWriter.
func (mw *multiWriter) Add(writer io.WriteCloser) {
	mw.Lock()
	mw.writers = append(mw.writers, writer)
	mw.Unlock()
}

// Write writes data into several writers and never returns error.
func (mw *multiWriter) Write(p []byte) (int, error) {
	mw.Lock()
	var evictIdx []int
	for n, w := range mw.writers {
		if _, err := w.Write(p); err != nil {
			log.With(nil).WithError(err).Debug("failed to write data")

			w.Close()
			evictIdx = append(evictIdx, n)
		}
	}

	for n, i := range evictIdx {
		mw.writers = append(mw.writers[:i-n], mw.writers[i-n+1:]...)
	}
	mw.Unlock()
	return len(p), nil
}

// Close closes all the writers and never returns error.
func (mw *multiWriter) Close() error {
	mw.Lock()
	for _, w := range mw.writers {
		w.Close()
	}
	mw.writers = nil
	mw.Unlock()
	return nil
}
