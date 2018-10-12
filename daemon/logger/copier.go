package logger

import (
	"bufio"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogCopier is used to copy data from stream and write it into LogDriver.
type LogCopier struct {
	sync.WaitGroup
	srcs map[string]io.Reader
	dst  LogDriver
}

// NewLogCopier creates copier for logger.
func NewLogCopier(dst LogDriver, srcs map[string]io.Reader) *LogCopier {
	return &LogCopier{
		srcs: srcs,
		dst:  dst,
	}
}

// StartCopy starts to read the data and write it into logger.
func (lc *LogCopier) StartCopy() {
	for source, r := range lc.srcs {
		lc.Add(1)
		go lc.copy(source, r)
	}
}

func (lc *LogCopier) copy(source string, reader io.Reader) {
	defer logrus.Debugf("finish %s stream type logcopy for %s", source, lc.dst.Name())
	defer lc.Done()

	var (
		bs  []byte
		err error

		firstPartial = true
		isPartial    bool
		createdTime  time.Time

		defaultBufSize = 16 * 1024
	)

	br := bufio.NewReaderSize(reader, defaultBufSize)
	for {
		bs, isPartial, err = br.ReadLine()
		if err != nil {
			if err != io.EOF {
				logrus.WithError(err).
					Errorf("failed to copy into %v-%v", lc.dst.Name(), source)
			}
			return
		}

		// NOTE: The partial content will share the same timestamp.
		if firstPartial {
			createdTime = time.Now().UTC()
		}

		if isPartial {
			firstPartial = false
		} else {
			firstPartial = true
			bs = append(bs, '\n')
		}

		logMessage := NewMessage()
		logMessage.Source = source
		logMessage.Line = bs
		logMessage.Timestamp = createdTime
		if err = lc.dst.WriteLogMessage(logMessage); err != nil {
			logrus.WithError(err).Errorf("failed to copy into %v-%v", lc.dst.Name(), source)
		}
	}
}
