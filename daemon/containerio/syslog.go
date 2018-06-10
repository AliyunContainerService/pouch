package containerio

import (
	"io"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/daemon/logger/syslog"
)

func init() {
	Register(func() Backend {
		return &syslogging{}
	})
}

type customWriter struct {
	w func(p []byte) (int, error)
}

func (cw *customWriter) Write(p []byte) (int, error) {
	return cw.w(p)
}

type syslogging struct {
	w *syslog.Syslog
}

func (s *syslogging) Init(opt *Option) error {
	w, err := syslog.NewSyslog(opt.info)
	if err != nil {
		return err
	}
	s.w = w
	return nil
}

func (s *syslogging) Name() string {
	return "syslog"
}

func (s *syslogging) In() io.Reader {
	return nil
}

func (s *syslogging) Out() io.Writer {
	return &customWriter{w: s.sourceWriteFunc("stdout")}
}

func (s *syslogging) Err() io.Writer {
	return &customWriter{w: s.sourceWriteFunc("stderr")}
}

func (s *syslogging) Close() error {
	return s.w.Close()
}

func (s *syslogging) sourceWriteFunc(source string) func([]byte) (int, error) {
	return func(p []byte) (int, error) {
		if len(p) == 0 {
			return 0, nil
		}

		msg := &logger.LogMessage{
			Source: source,
			Line:   p,
		}

		if err := s.w.WriteLogMessage(msg); err != nil {
			return 0, err
		}
		return len(p), nil
	}
}
