package logger

// LogDriver represents any kind of log drivers, such as jsonfile, syslog.
type LogDriver interface {
	Name() string

	WriteLogMessage(msg *LogMessage) error

	Close() error
}
