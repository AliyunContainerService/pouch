package log

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type key int

const (
	loggerKey key = iota
)

// Init initializes log Level and log format.
func Init(debug bool) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	logrus.SetFormatter(formatter)
}

// NewContext returns new log entry, if context has old entry, it will be overwrite
func NewContext(ctx context.Context, fields map[string]interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	logger := logrus.StandardLogger().WithFields(fields)
	return context.WithValue(ctx, loggerKey, logger)
}

// AddFields merges new fields with context logger's fields, and update new logger into context
// if ctx is nil, it is same with WithFields().
func AddFields(ctx context.Context, fields map[string]interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	logger, ok := ctx.Value(loggerKey).(*logrus.Entry)
	if !ok || fields == nil {
		logger = logrus.StandardLogger().WithFields(fields)
	} else {
		logger = logger.WithFields(fields)
	}

	// update logger to context
	return context.WithValue(ctx, loggerKey, logger)
}

// With returns log entry from context
func With(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logrus.StandardLogger().WithFields(nil)
	}

	logger, ok := ctx.Value(loggerKey).(*logrus.Entry)
	if !ok {
		return logrus.StandardLogger().WithFields(nil)
	}
	return logger
}

// WithFields merges new fields with context logger's fields, but don't update into context
func WithFields(ctx context.Context, fields map[string]interface{}) *logrus.Entry {
	if ctx == nil {
		return logrus.StandardLogger().WithFields(fields)
	}

	logger, ok := ctx.Value(loggerKey).(*logrus.Entry)
	if !ok {
		logger = logrus.StandardLogger().WithFields(fields)
	} else {
		logger = logger.WithFields(fields)
	}

	return logger
}
