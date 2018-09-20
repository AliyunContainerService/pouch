package containerio

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/alibaba/pouch/pkg/ioutils"

	containerdio "github.com/containerd/containerd/cio"
	"github.com/containerd/fifo"
)

// cio is a basic container IO implementation.
type cio struct {
	config containerdio.Config

	closer *wgCloser
}

func (c *cio) Config() containerdio.Config {
	return c.config
}

func (c *cio) Cancel() {
	if c.closer == nil {
		return
	}
	c.closer.Cancel()
}

func (c *cio) Wait() {
	if c.closer == nil {
		return
	}
	c.closer.Wait()
}

func (c *cio) Close() error {
	if c.closer == nil {
		return nil
	}
	return c.closer.Close()
}

// NewFifos returns a new set of fifos for the task
func NewFifos(id string, stdin bool) (*containerdio.FIFOSet, error) {
	root := "/run/containerd/fifo"
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	dir, err := ioutil.TempDir(root, "")
	if err != nil {
		return nil, err
	}
	fifos := &containerdio.FIFOSet{
		Dir: dir,
		In:  filepath.Join(dir, id+"-stdin"),
		Out: filepath.Join(dir, id+"-stdout"),
		Err: filepath.Join(dir, id+"-stderr"),
	}

	if !stdin {
		fifos.In = ""
	}

	return fifos, nil
}

// Stream is used to configure the stream from client side
type Stream struct {
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	Terminal bool
}

type wgCloser struct {
	wg     *sync.WaitGroup
	dir    string
	set    []io.Closer
	cancel context.CancelFunc
}

func (g *wgCloser) Wait() {
	g.wg.Wait()
}

func (g *wgCloser) Close() error {
	for _, f := range g.set {
		f.Close()
	}
	if g.dir != "" {
		return os.RemoveAll(g.dir)
	}
	return nil
}

func (g *wgCloser) Cancel() {
	g.cancel()
}

type pipes struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

func openPipes(ctx context.Context, fifos *containerdio.FIFOSet) (pipes, error) {
	var (
		err error
		p   pipes
	)

	if fifos.In != "" {
		if p.Stdin, err = fifo.OpenFifo(ctx, fifos.In, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
			return p, err
		}

		defer func() {
			if err != nil && p.Stdin != nil {
				p.Stdin.Close()
			}
		}()
	}

	if p.Stdout, err = fifo.OpenFifo(ctx, fifos.Out, syscall.O_RDONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
		return p, err
	}
	defer func() {
		if err != nil && p.Stdout != nil {
			p.Stdout.Close()
		}
	}()

	if p.Stderr, err = fifo.OpenFifo(ctx, fifos.Err, syscall.O_RDONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
		return p, err
	}
	return p, nil
}

func copyStream(fifoset *containerdio.FIFOSet, stream *Stream, closeStdin func() error) (*wgCloser, error) {
	// if fifos directory is not exist, create fifo will fails,
	// also in case of fifo directory lost in container recovery process.
	if _, err := os.Stat(fifoset.Dir); err != nil && os.IsNotExist(err) {
		os.MkdirAll(fifoset.Dir, 0700)
	}

	var (
		err     error
		pipe    pipes
		closers []io.Closer
		wg      = &sync.WaitGroup{}
	)

	ctx, cancel := context.WithCancel(context.Background())
	if pipe, err = openPipes(ctx, fifoset); err != nil {
		cancel()
		return nil, err
	}

	defer func() {
		if err != nil {
			for _, f := range closers {
				f.Close()
			}
			cancel()
		}
	}()

	if pipe.Stdin != nil {
		if !fifoset.Terminal && closeStdin != nil {
			var (
				closeStdinOnce sync.Once
				werr           error
			)

			oldStdin := pipe.Stdin
			pipe.Stdin = ioutils.NewWriteCloserWrapper(oldStdin, func() error {
				closeStdinOnce.Do(func() {
					werr = oldStdin.Close()
					if werr != nil {
						return
					}

					werr = closeStdin()
				})
				return werr
			})

		}

		closers = append(closers, pipe.Stdin)
		go func(w io.WriteCloser) {
			io.Copy(w, stream.Stdin)
			w.Close()
		}(pipe.Stdin)
	}

	wg.Add(1)
	closers = append(closers, pipe.Stdout)
	go func(r io.ReadCloser) {
		defer wg.Done()
		io.Copy(stream.Stdout, r)
		r.Close()
	}(pipe.Stdout)

	if !fifoset.Terminal {
		wg.Add(1)
		closers = append(closers, pipe.Stderr)
		go func(r io.ReadCloser) {
			defer wg.Done()
			io.Copy(stream.Stderr, r)
			r.Close()
		}(pipe.Stderr)
	}

	return &wgCloser{
		wg:     wg,
		dir:    fifoset.Dir,
		set:    closers,
		cancel: cancel,
	}, nil
}

// NewIOWithTerminal creates a new io set with the provied io.Reader/Writers for use with a terminal
func NewIOWithTerminal(stream *Stream, enableStdin bool, closeStdin func() error) containerdio.Creation {
	return func(id string) (containerdio.IO, error) {
		var (
			fifoset *containerdio.FIFOSet
			err     error
		)

		fifoset, err = NewFifos(id, enableStdin)
		if err != nil {
			return nil, err
		}
		fifoset.Terminal = stream.Terminal

		defer func() {
			if err != nil && fifoset.Dir != "" {
				os.RemoveAll(fifoset.Dir)
			}
		}()

		cfg := containerdio.Config{
			Terminal: fifoset.Terminal,
			Stdout:   fifoset.Out,
			Stderr:   fifoset.Err,
			Stdin:    fifoset.In,
		}

		i := &cio{config: cfg}
		if i.closer, err = copyStream(fifoset, stream, closeStdin); err != nil {
			return nil, err
		}
		return i, nil
	}
}

// WithAttach attaches the existing io for a task to the provided io.Reader/Writers
func WithAttach(stream *Stream) containerdio.Attach {
	return func(fifoset *containerdio.FIFOSet) (containerdio.IO, error) {
		var err error
		if fifoset == nil {
			return nil, fmt.Errorf("cannot attach to existing fifos")
		}

		cfg := containerdio.Config{
			Terminal: fifoset.Terminal,
			Stdout:   fifoset.Out,
			Stderr:   fifoset.Err,
			Stdin:    fifoset.In,
		}

		i := &cio{config: cfg}
		// FIXME(fuweid): should we add closeStdin for recovery container
		// like NewIOWithTerminal? if we don't set closeStdin, the attach
		// action can't close the IO.
		if i.closer, err = copyStream(fifoset, stream, nil); err != nil {
			return nil, err
		}
		return i, nil
	}
}
