package syslog

import (
	"crypto/tls"
	"os"
	"strings"
	"sync"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/daemon/logger/loggerutils"

	"github.com/RackSec/srslog"
)

// Syslog writes the log data into syslog.
type Syslog struct {
	mu sync.RWMutex

	opt  *options
	conn serverConn
}

type options struct {
	tag       string
	proto     string
	address   string
	hostname  string
	priority  Priority
	formatter Formatter
	framer    Framer
	tlsCfg    *tls.Config
}

func defaultOptions() *options {
	return &options{
		priority: defaultSyslogPriority,
	}
}

// Init return the Syslog log driver.
func Init(info logger.Info) (logger.LogDriver, error) {
	return NewSyslog(info)
}

// NewSyslog returns new Syslog based on the log config.
func NewSyslog(info logger.Info) (*Syslog, error) {
	opt, err := parseOptions(info)
	if err != nil {
		return nil, err
	}

	opt.hostname, _ = os.Hostname()
	return &Syslog{
		opt:  opt,
		conn: nil,
	}, nil
}

// Name return the log driver's name.
func (s *Syslog) Name() string {
	return "syslog"
}

// WriteLogMessage will write the LogMessage.
func (s *Syslog) WriteLogMessage(msg *logger.LogMessage) error {
	line := string(msg.Line)
	source := msg.Source
	logger.PutMessage(msg)

	if source == "stderr" {
		return s.logError(line)
	}
	return s.logInfo(line)
}

// Close closes the Syslog.
func (s *Syslog) Close() error {
	var err error
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		err = s.conn.close()
		s.conn = nil
	}
	return err
}

// logInfo logs a content with severity LOG_INFO.
func (s *Syslog) logInfo(content string) error {
	_, err := s.writeAndRetry(srslog.LOG_INFO, content)
	return err
}

// logError logs a content with severity LOG_ERR.
func (s *Syslog) logError(content string) error {
	_, err := s.writeAndRetry(srslog.LOG_ERR, content)
	return err
}

// writeAndRetry takes a severity and the content to write.
//
// NOTE: Any facility passed to it as part of the severity Priority will be ignored.
func (s *Syslog) writeAndRetry(severity Priority, content string) (int, error) {
	p := (s.opt.priority & facilityMask) | (severity & severityMask)

	conn := s.getConn()
	if conn != nil {
		if n, err := s.write(conn, p, content); err == nil {
			return n, nil
		}
	}

	var err error
	if conn, err = s.connect(); err != nil {
		return 0, err
	}
	return s.write(conn, p, content)
}

// write writes a syslog formatted string.
func (s *Syslog) write(conn serverConn, p Priority, content string) (int, error) {
	// ensure it ends with a \n
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	err := conn.writeString(s.opt.framer, s.opt.formatter, p, s.opt.hostname, s.opt.tag, content)
	if err != nil {
		return 0, err
	}

	return len(content), nil
}

// connect uses current option to connect the remote host.
func (s *Syslog) connect() (serverConn, error) {
	sc, hostname, err := makeDialer(s.opt.proto, s.opt.address, s.opt.tlsCfg)
	if err != nil {
		return nil, err
	}

	s.setConn(sc, hostname)
	return sc, nil
}

// getConn returns the current serverConn.
func (s *Syslog) getConn() serverConn {
	s.mu.RLock()
	c := s.conn
	s.mu.RUnlock()
	return c
}

// setConn updates the connection.
//
// NOTE: the Syslog takes lazy mode for connection. It might have more goroutines
// which try to connect the same remote host. If there is no close existing
// connection, it will be connection leak.
func (s *Syslog) setConn(c serverConn, hostname string) {
	s.mu.Lock()
	if s.conn != nil {
		s.conn.close()
	}

	s.conn = c
	if s.opt.hostname == "" {
		s.opt.hostname = hostname
	}
	s.mu.Unlock()
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
