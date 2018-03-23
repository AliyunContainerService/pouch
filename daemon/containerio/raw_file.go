package containerio

import (
	"io"
	"os"
	"path/filepath"
	"syscall"
)

func init() {
	Register(func() Backend {
		return &rawFile{}
	})
}

type rawFile struct {
	file   *os.File
	closed bool
}

func (r *rawFile) Name() string {
	return "raw-file"
}

func (r *rawFile) Init(opt *Option) error {
	path := filepath.Join(opt.rootDir, "raw-file.out")
	f, err := os.OpenFile(path, syscall.O_CREAT|syscall.O_APPEND|syscall.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	r.file = f
	return nil
}

func (r *rawFile) Out() io.Writer {
	return r.file
}

func (r *rawFile) In() io.Reader {
	return r.file
}

func (r *rawFile) Err() io.Writer {
	return r.file
}

func (r *rawFile) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	return r.file.Close()
}
