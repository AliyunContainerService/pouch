package jsonfile

import (
	"fmt"
	"os"
	"sync"

	"github.com/alibaba/pouch/daemon/logger"
)

//MarshalFunc is the function of marshal the logMessage
type MarshalFunc func(message *logger.LogMessage) ([]byte, error)

// JSONLogFile is uses to log the container's stdout and stderr.
type JSONLogFile struct {
	mu sync.Mutex

	f           *os.File
	perms       os.FileMode
	closed      bool
	marshalFunc MarshalFunc
}

// NewJSONLogFile returns new JSONLogFile instance.
func NewJSONLogFile(logPath string, perms os.FileMode, marshalFunc MarshalFunc) (*JSONLogFile, error) {
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, perms)
	if err != nil {
		return nil, err
	}

	return &JSONLogFile{
		f:           f,
		perms:       perms,
		closed:      false,
		marshalFunc: marshalFunc,
	}, nil
}

// WriteLogMessage will write the LogMessage into the file.
func (lf *JSONLogFile) WriteLogMessage(msg *logger.LogMessage) error {
	b, err := lf.marshalFunc(msg)
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

var validLogOpt = []string{"max-file", "max-size", "compress", "labels", "env", "env-regex", "tag"}

// ValidateLogOpt validate log options for json-file log driver
func ValidateLogOpt(cfg map[string]string) error {
	for key := range cfg {
		isValid := false
		for _, opt := range validLogOpt {
			if key == opt {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("unknown log opt '%s' for json-file log driver", key)
		}
	}
	return nil
}
