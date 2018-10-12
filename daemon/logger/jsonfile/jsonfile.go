package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alibaba/pouch/daemon/logger"
)

var jsonFilePathName = "json.log"

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

// Init initializes the jsonfile log driver.
func Init(info logger.Info) (logger.LogDriver, error) {
	if _, err := os.Stat(info.ContainerRootDir); err != nil {
		return nil, err
	}

	logPath := filepath.Join(info.ContainerRootDir, jsonFilePathName)

	attrs, err := info.ExtraAttributes(nil)
	if err != nil {
		return nil, err
	}

	var extra []byte
	if len(attrs) > 0 {
		var err error
		extra, err = json.Marshal(attrs)
		if err != nil {
			return nil, err
		}
	}

	return NewJSONLogFile(logPath, 0644, func(msg *logger.LogMessage) ([]byte, error) {
		return Marshal(msg, extra)
	})
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

// Name return the log driver's name.
func (lf *JSONLogFile) Name() string {
	return "json-file"
}

// WriteLogMessage will write the LogMessage into the file.
func (lf *JSONLogFile) WriteLogMessage(msg *logger.LogMessage) error {
	b, err := lf.marshalFunc(msg)
	if err != nil {
		return err
	}
	logger.PutMessage(msg)

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
