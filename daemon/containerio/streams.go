package containerio

import (
	"io"

	"github.com/alibaba/pouch/cri/stream/remotecommand"
)

func init() {
	Register(func() Backend {
		return &streamIO{}
	})
}

type streamIO struct {
	streams *remotecommand.Streams
	closed  bool
}

func (s *streamIO) Name() string {
	return "streams"
}

func (s *streamIO) Init(opt *Option) error {
	s.streams = opt.streams

	return nil
}

func (s *streamIO) Out() io.Writer {
	return s.streams.StdoutStream
}

func (s *streamIO) In() io.Reader {
	return s.streams.StdinStream
}

func (s *streamIO) Close() error {
	if s.closed {
		return nil
	}

	for _, closer := range []io.Closer{
		s.streams.StdinStream,
		s.streams.StdoutStream,
		s.streams.StderrStream,
	} {
		if closer != nil {
			closer.Close()
		}
	}

	if s.streams.StreamCh != nil {
		s.streams.StreamCh <- struct{}{}
	}

	s.closed = true

	return nil
}
