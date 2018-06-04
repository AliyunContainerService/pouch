package logger

import (
	"sync"
	"time"
)

// LogMessage represents the log message in the container json log.
type LogMessage struct {
	Source    string    // Source means stdin, stdout or stderr
	Line      []byte    // Line means the log content, but it maybe partial
	Timestamp time.Time // Timestamp means the created time of line

	Err error
}

// LogWatcher is used to pass the log message to the reader.
type LogWatcher struct {
	Msgs chan *LogMessage
	Err  chan error

	closeOnce     sync.Once
	closeNotifier chan struct{}
}

const defaultBufSize = 2048

// NewLogWatcher returns new LogWatcher.
func NewLogWatcher() *LogWatcher {
	return &LogWatcher{
		Msgs:          make(chan *LogMessage, defaultBufSize),
		Err:           make(chan error, 1),
		closeNotifier: make(chan struct{}),
	}
}

// Close closes LogWatcher.
func (w *LogWatcher) Close() {
	w.closeOnce.Do(func() {
		close(w.closeNotifier)
	})
}

// WatchClose returns the close notifier.
func (w *LogWatcher) WatchClose() <-chan struct{} {
	return w.closeNotifier
}

// ReadConfig is used to
type ReadConfig struct {
	Since  time.Time
	Until  time.Time
	Tail   int
	Follow bool
}
