package syslog

import (
	"crypto/tls"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/daemon/logger/loggerutils"

	"github.com/RackSec/srslog"
)

// Syslog writes the log data into syslog.
type Syslog struct {
	writer *srslog.Writer
}

type options struct {
	tag       string
	proto     string
	address   string
	priority  srslog.Priority
	formatter srslog.Formatter
	framer    srslog.Framer
	tlsCfg    *tls.Config
}

func defaultOptions() *options {
	return &options{
		priority: defaultSyslogPriority,
	}
}

// NewSyslog returns new Syslog based on the log config.
func NewSyslog(info logger.Info) (*Syslog, error) {
	opts, err := parseOptions(info)
	if err != nil {
		return nil, err
	}

	var w *srslog.Writer
	if opts.proto == secureProto {
		w, err = srslog.DialWithTLSConfig(opts.proto, opts.address, opts.priority, opts.tag, opts.tlsCfg)
	} else {
		w, err = srslog.Dial(opts.proto, opts.address, opts.priority, opts.tag)
	}

	if err != nil {
		return nil, err
	}

	w.SetFormatter(opts.formatter)
	w.SetFramer(opts.framer)
	return &Syslog{writer: w}, nil
}

// WriteLogMessage will write the LogMessage.
func (s *Syslog) WriteLogMessage(msg *logger.LogMessage) error {
	line := string(msg.Line)
	if msg.Source == "stderr" {
		return s.writer.Err(line)
	}
	return s.writer.Info(line)
}

// Close closes the Syslog.
func (s *Syslog) Close() error {
	return s.writer.Close()
}

// parseOptions parses the log config into options.
func parseOptions(info logger.Info) (*options, error) {
	var err error
	opts := defaultOptions()

	opts.priority, err = parseFacility(info.LogConfig["syslog-facility"])
	if err != nil {
		return nil, err
	}

	opts.tag, err = loggerutils.GenerateLogTag(info, defaultTagTemplate)
	if err != nil {
		return nil, err
	}

	opts.proto, opts.address, err = parseTargetAddress(info.LogConfig["syslog-address"])
	if err != nil {
		return nil, err
	}

	if opts.proto == secureProto {
		opts.tlsCfg, err = parseTLSConfig(info)
		if err != nil {
			return nil, err
		}
	}

	opts.formatter, opts.framer, err = parseLogFormat(info.LogConfig["syslog-format"], opts.proto)
	if err != nil {
		return nil, err
	}
	return opts, nil
}
