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

	/*
		NOTE(tanghuamin): for compatible docker 1.12
		if !stdin {
			fifos.In = ""
		}
	*/

	return fifos, nil
}

type ioSet struct {
	in       io.Reader
	out, err io.Writer
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

func copyIO(fifos *containerdio.FIFOSet, ioset *ioSet, tty bool) (_ *wgCloser, err error) {
	var (
		f           io.ReadWriteCloser
		set         []io.Closer
		ctx, cancel = context.WithCancel(context.Background())
		wg          = &sync.WaitGroup{}
	)
	defer func() {
		if err != nil {
			for _, f := range set {
				f.Close()
			}
			cancel()
		}
	}()

	if fifos.In != "" {
		if f, err = fifo.OpenFifo(ctx, fifos.In, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
			return nil, err
		}
		set = append(set, f)
		go func(w io.WriteCloser) {
			io.Copy(w, ioset.in)
			w.Close()
		}(f)
	}

	if f, err = fifo.OpenFifo(ctx, fifos.Out, syscall.O_RDONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
		return nil, err
	}
	set = append(set, f)
	wg.Add(1)
	go func(r io.ReadCloser) {
		io.Copy(ioset.out, r)
		r.Close()
		wg.Done()
	}(f)

	if f, err = fifo.OpenFifo(ctx, fifos.Err, syscall.O_RDONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0700); err != nil {
		return nil, err
	}
	set = append(set, f)

	if !tty {
		wg.Add(1)
		go func(r io.ReadCloser) {
			io.Copy(ioset.err, r)
			r.Close()
			wg.Done()
		}(f)
	}
	return &wgCloser{
		wg:     wg,
		dir:    fifos.Dir,
		set:    set,
		cancel: cancel,
	}, nil
}

// NewIOWithTerminal creates a new io set with the provied io.Reader/Writers for use with a terminal
func NewIOWithTerminal(stdin io.Reader, stdout, stderr io.Writer, terminal bool, stdinEnable bool) containerdio.Creation {
	return func(id string) (_ containerdio.IO, err error) {
		paths, err := NewFifos(id, stdinEnable)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil && paths.Dir != "" {
				os.RemoveAll(paths.Dir)
			}
		}()
		cfg := containerdio.Config{
			Terminal: terminal,
			Stdout:   paths.Out,
			Stderr:   paths.Err,
			Stdin:    paths.In,
		}
		i := &cio{config: cfg}
		set := &ioSet{
			in:  stdin,
			out: stdout,
			err: stderr,
		}
		closer, err := copyIO(paths, set, cfg.Terminal)
		if err != nil {
			return nil, err
		}
		i.closer = closer
		return i, nil
	}
}

// WithAttach attaches the existing io for a task to the provided io.Reader/Writers
func WithAttach(stdin io.Reader, stdout, stderr io.Writer) containerdio.Attach {
	return func(paths *containerdio.FIFOSet) (containerdio.IO, error) {
		if paths == nil {
			return nil, fmt.Errorf("cannot attach to existing fifos")
		}
		cfg := containerdio.Config{
			Terminal: paths.Terminal,
			Stdout:   paths.Out,
			Stderr:   paths.Err,
			Stdin:    paths.In,
		}
		i := &cio{config: cfg}
		set := &ioSet{
			in:  stdin,
			out: stdout,
			err: stderr,
		}
		closer, err := copyIO(paths, set, cfg.Terminal)
		if err != nil {
			return nil, err
		}
		i.closer = closer
		return i, nil
	}
}
