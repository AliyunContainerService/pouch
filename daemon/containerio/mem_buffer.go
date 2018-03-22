package containerio

import (
	"bytes"
	"io"
)

func init() {
	Register(func() Backend {
		return &memBuffer{}
	})
}

type memBuffer struct {
	buffer *bytes.Buffer
}

func (b *memBuffer) Name() string {
	return "memBuffer"
}

func (b *memBuffer) Init(opt *Option) error {
	b.buffer = opt.memBuffer
	return nil
}

func (b *memBuffer) Out() io.Writer {
	return b.buffer
}

func (b *memBuffer) In() io.Reader {
	return b.buffer
}

func (b *memBuffer) Err() io.Writer {
	return b.buffer
}

func (b *memBuffer) Close() error {
	// Don't need to close bytes.Buffer.
	return nil
}
