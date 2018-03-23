package containerio

import (
	"io"
)

func init() {
	Register(func() Backend {
		return &discardIO{}
	})
}

type discardIO struct{}

func (d *discardIO) Name() string {
	return "discard"
}

func (d *discardIO) Init(opt *Option) error {
	return nil
}

func (d *discardIO) Out() io.Writer {
	return d
}

func (d *discardIO) Err() io.Writer {
	return d
}

func (d *discardIO) In() io.Reader {
	return d
}

func (d *discardIO) Close() error {
	return nil
}

func (d *discardIO) Write(data []byte) (int, error) {
	return len(data), nil
}

func (d *discardIO) Read(p []byte) (int, error) {
	block := make(chan struct{})
	<-block
	return 0, nil
}
