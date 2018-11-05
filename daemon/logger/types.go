package logger

// LogMode indicates available logging modes.
type LogMode string

const (
	// LogModeDefault default to unuse buffer to make logs blocking.
	LogModeDefault = ""
	// LogModeBlocking means to unuse buffer to make logs blocking.
	LogModeBlocking LogMode = "blocking"
	// LogModeNonBlock means to use buffer to make logs non blocking.
	LogModeNonBlock LogMode = "non-blocking"
)

// LogDriver represents any kind of log drivers, such as jsonfile, syslog.
type LogDriver interface {
	Name() string

	WriteLogMessage(msg *LogMessage) error

	Close() error
}
