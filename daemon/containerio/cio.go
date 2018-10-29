package containerio

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/containerd/containerd/cio"
	"github.com/containerd/fifo"
	"github.com/sirupsen/logrus"
)

// CioFIFOSet holds the fifo pipe for containerd shim.
//
// It will be replaced by the containerd@v1.2
type CioFIFOSet struct {
	cio.Config
	closeFn func() error
}

// Close will remove all the relatived files.
func (c *CioFIFOSet) Close() error {
	if c.closeFn != nil {
		return c.closeFn()
	}
	return nil
}

// NewCioFIFOSet prepares fifo files.
func NewCioFIFOSet(processID string, withStdin bool, withTerminal bool) (*CioFIFOSet, error) {
	root := "/run/containerd/fifo"
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}

	fifoDir, err := ioutil.TempDir(root, "")
	if err != nil {
		return nil, err
	}

	cfg := cio.Config{
		Terminal: withTerminal,
		Stdout:   filepath.Join(fifoDir, processID+"-stdout"),
	}

	if withStdin {
		cfg.Stdin = filepath.Join(fifoDir, processID+"-stdin")
	}

	if !withTerminal {
		cfg.Stderr = filepath.Join(fifoDir, processID+"-stderr")
	}

	closeFn := func() error {
		err := os.RemoveAll(fifoDir)
		if err != nil {
			logrus.WithError(err).Warnf("failed to remove process(id=%v) fifo dir", processID)
		}
		return err
	}

	return &CioFIFOSet{
		Config:  cfg,
		closeFn: closeFn,
	}, nil
}

// TODO(fuweid): containerIO will be removed when update vendor to containerd@v1.2.
type containerIO struct {
	config  cio.Config
	wg      *sync.WaitGroup
	closers []io.Closer
	cancel  context.CancelFunc
}

func (c *containerIO) Config() cio.Config {
	return c.config
}

func (c *containerIO) Wait() {
	if c.wg != nil {
		c.wg.Wait()
	}
}

func (c *containerIO) Close() error {
	var lastErr error
	for _, closer := range c.closers {
		if closer == nil {
			continue
		}
		if err := closer.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (c *containerIO) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}
}

// TODO(fuweid): pipes will be removed when update vendor to containerd@v1.2.
type pipes struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

func (p *pipes) Closers() []io.Closer {
	return []io.Closer{p.Stdout, p.Stderr, p.Stdin}
}

// DirectIO allows task IO to be handled externally by the caller.
//
// TODO(fuweid): DirectIO will be removed when update vendor to containerd@v1.2.
type DirectIO struct {
	pipes
	containerIO
}

var _ cio.IO = &DirectIO{}

// NewDirectIO returns an IO implementation that exposes the IO streams as
// io.ReadCloser and io.WriteCloser.
//
// TODO(fuweid): NewDirectIO will be removed when update vendor to containerd@v1.2.
func NewDirectIO(ctx context.Context, fifos *CioFIFOSet) (*DirectIO, error) {
	ctx, cancel := context.WithCancel(ctx)
	pipes, err := openPipes(ctx, fifos)
	if err != nil {
		cancel()
		return nil, err
	}

	return &DirectIO{
		pipes: pipes,
		containerIO: containerIO{
			config:  fifos.Config,
			closers: append(pipes.Closers(), fifos),
			cancel:  cancel,
		},
	}, nil
}

func openPipes(ctx context.Context, fifos *CioFIFOSet) (_ pipes, err0 error) {
	var (
		err error
		p   pipes
	)

	defer func() {
		if err0 != nil {
			fifos.Close()
		}
	}()

	if fifos.Stdin != "" {
		if p.Stdin, err = fifo.OpenFifo(ctx, fifos.Stdin, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
			return p, err
		}

		defer func() {
			if err != nil && p.Stdin != nil {
				p.Stdin.Close()
			}
		}()
	}

	if fifos.Stdout != "" {
		if p.Stdout, err = fifo.OpenFifo(ctx, fifos.Stdout, syscall.O_RDONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
			return p, err
		}
		defer func() {
			if err != nil && p.Stdout != nil {
				p.Stdout.Close()
			}
		}()
	}

	if fifos.Stderr != "" {
		if p.Stderr, err = fifo.OpenFifo(ctx, fifos.Stderr, syscall.O_RDONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
			return p, err
		}
	}
	return p, nil
}
