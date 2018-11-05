package logbuffer

import (
	"github.com/alibaba/pouch/daemon/logger"

	"github.com/sirupsen/logrus"
)

// LogBuffer is uses to cache the container's logs with ringBuffer.
type LogBuffer struct {
	ringBuffer *RingBuffer
	logger     logger.LogDriver
}

// NewLogBuffer return a new BufferLog.
func NewLogBuffer(logDriver logger.LogDriver, maxBytes int64) (logger.LogDriver, error) {
	bl := &LogBuffer{
		logger:     logDriver,
		ringBuffer: NewRingBuffer(maxBytes),
	}

	// use a goroutine to write logs continuously with specified log driver
	go bl.run()
	return bl, nil
}

// Name return the log driver's name.
func (bl *LogBuffer) Name() string {
	return bl.logger.Name()
}

// WriteLogMessage will write the LogMessage to the ringBuffer.
func (bl *LogBuffer) WriteLogMessage(msg *logger.LogMessage) error {
	return bl.ringBuffer.Push(msg)
}

// Close close the ringBuffer and drain the messages.
func (bl *LogBuffer) Close() error {
	bl.ringBuffer.Close()
	for _, msg := range bl.ringBuffer.Drain() {
		if err := bl.logger.WriteLogMessage(msg); err != nil {
			logrus.Debugf("failed to write log %v when closing with log driver %s", msg, bl.logger.Name())
		}
	}

	return bl.logger.Close()
}

// write logs continuously with specified log driver from ringBuffer.
func (bl *LogBuffer) run() {
	for {
		msg, err := bl.ringBuffer.Pop()
		if err != nil {
			return
		}

		if err := bl.logger.WriteLogMessage(msg); err != nil {
			logrus.Debugf("failed to write log %v with log driver %s", msg, bl.logger.Name())
		}
	}
}
