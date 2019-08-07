package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/pkg/bytefmt"
)

const defaultMaxSize = uint64(100 * 1024 * 1024)
const defaultMaxFile = 2

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
	maxSize     uint64 // maximum size of log in byte
	currentSize uint64 // current size of the latest log in byte
	maxFile     int    // maximum number of logs
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

	return NewJSONLogFile(logPath, 0644, info.LogConfig, func(msg *logger.LogMessage) ([]byte, error) {
		return Marshal(msg, extra)
	})
}

// NewJSONLogFile returns new JSONLogFile instance.
func NewJSONLogFile(logPath string, perms os.FileMode, logConfig map[string]string, marshalFunc MarshalFunc) (*JSONLogFile, error) {
	var (
		currentSize uint64
		maxSize     = defaultMaxSize
		maxFiles    = defaultMaxFile
	)
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, perms)
	if err != nil {
		return nil, err
	}

	if logConfig != nil {
		size, err := f.Seek(0, os.SEEK_END)
		if err != nil {
			return nil, err
		}
		currentSize = uint64(size)
		if maxSizeString, ok := logConfig["max-size"]; ok {
			maxSize, err = bytefmt.ToBytes(maxSizeString)
			if err != nil {
				return nil, err
			}
		}
		if maxFileString, ok := logConfig["max-file"]; ok {
			maxFiles, err = strconv.Atoi(maxFileString)
			if err != nil {
				return nil, err
			}
			if maxFiles < 1 {
				return nil, fmt.Errorf("max-file cannot be less than 1")
			}
		}
	}

	return &JSONLogFile{
		f:           f,
		perms:       perms,
		closed:      false,
		marshalFunc: marshalFunc,
		maxSize:     maxSize,
		currentSize: currentSize,
		maxFile:     maxFiles,
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

	lf.mu.Lock()
	defer lf.mu.Unlock()
	if err = lf.checkRotate(); err != nil {
		return err
	}

	n, err := lf.f.Write(b)
	if err == nil {
		lf.currentSize += uint64(n)
	}
	return err
}

// checkRotate rotates logs according to maxSize and maxFile parameters
// TODO: after rotating logs, a notice should be made to the log reader
func (lf *JSONLogFile) checkRotate() error {
	if lf.maxSize == 0 || lf.currentSize < lf.maxSize {
		// no need to rotate
		return nil
	}

	logName := lf.f.Name()
	// step1. close current log file
	if err := lf.f.Close(); err != nil {
		return err
	}
	// step2. rotate logs. move x.log.(n-1) to x.log.n
	if err := rotate(logName, lf.maxFile); err != nil {
		return err
	}
	// step3. reopen new log file with the same name
	newfile, err := os.OpenFile(logName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	lf.f = newfile
	lf.currentSize = 0

	return nil
}

func rotate(logName string, maxFiles int) error {
	if maxFiles < 2 {
		return nil
	}
	for i := maxFiles - 1; i > 1; i-- {
		newName := logName + "." + strconv.Itoa(i)
		oldName := logName + "." + strconv.Itoa(i-1)
		if err := os.Rename(oldName, newName); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	if err := os.Rename(logName, logName+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
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
