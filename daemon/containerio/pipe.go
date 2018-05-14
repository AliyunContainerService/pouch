package containerio

import (
	"io"
)

func init() {
	Register(func() Backend {
		return &pipe{}
	})
}

type pipe struct {
	pipeWriter *io.PipeWriter
}

func (p *pipe) Name() string {
	return "pipe"
}

func (p *pipe) Init(opt *Option) error {
	p.pipeWriter = opt.pipe
	return nil
}

func (p *pipe) Out() io.Writer {
	return p.pipeWriter
}

func (p *pipe) In() io.Reader {
	return nil
}

func (p *pipe) Err() io.Writer {
	return p.pipeWriter
}

func (p *pipe) Close() error {
	return p.pipeWriter.Close()

}
