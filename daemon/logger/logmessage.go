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
	Attrs     map[string]string
	Err       error
}

var messagePool = &sync.Pool{New: func() interface{} { return &LogMessage{Line: make([]byte, 0, 256)} }}

// NewMessage returns a new message from the message sync.Pool
func NewMessage() *LogMessage {
	return messagePool.Get().(*LogMessage)
}

// PutMessage puts the specified message back n the message pool.
// The message fields are reset before putting into the pool.
func PutMessage(msg *LogMessage) {
	msg.reset()
	messagePool.Put(msg)
}

// reset sets the message back to default values
// This is used when putting a message back into the message pool.
func (m *LogMessage) reset() {
	m.Line = m.Line[:0]
	m.Source = ""
	m.Err = nil
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
	Since   time.Time
	Until   time.Time
	Tail    int
	Follow  bool
	Details bool
}
