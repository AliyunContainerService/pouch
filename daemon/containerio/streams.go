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
	for _, closer := range []io.Closer{
		s.streams.StdinStream,
		s.streams.StdoutStream,
		s.streams.StderrStream,
	}{
		if closer != nil {
			closer.Close()
		}
	}

	return nil
}

func (s *streamIO) Write(data []byte) (int, error) {
	return s.streams.StdoutStream.Write(data)
}

func (s *streamIO) Read(p []byte) (int, error) {
	return s.streams.StdinStream.Read(p)
}
