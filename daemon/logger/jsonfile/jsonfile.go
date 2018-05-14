package jsonfile

import (
	"os"
	"sync"

	"github.com/alibaba/pouch/daemon/logger"
)

// JSONLogFile is uses to log the container's stdout and stderr.
type JSONLogFile struct {
	mu sync.Mutex

	f      *os.File
	perms  os.FileMode
	closed bool
}

// NewJSONLogFile returns new JSONLogFile instance.
func NewJSONLogFile(logPath string, perms os.FileMode) (*JSONLogFile, error) {
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, perms)
	if err != nil {
		return nil, err
	}

	return &JSONLogFile{
		f:      f,
		perms:  perms,
		closed: false,
	}, nil
}

// WriteLogMessage will write the LogMessage into the file.
func (lf *JSONLogFile) WriteLogMessage(msg *logger.LogMessage) error {
	b, err := marshal(msg)
	if err != nil {
		return err
	}

	lf.mu.Lock()
	defer lf.mu.Unlock()
	_, err = lf.f.Write(b)
	return err
}

// Close closes the file.
func (lf *JSONLogFile) Close() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.closed {
		return nil
	}

	if err := lf.f.Close(); err != nil {
		return err
	}
	lf.closed = true
	return nil
}
