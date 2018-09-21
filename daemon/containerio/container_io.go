package containerio

import (
	"fmt"
	"io"

	"github.com/alibaba/pouch/pkg/ringbuffer"

	"github.com/sirupsen/logrus"
)

const (
	stdout stdioType = iota
	stderr
	stdin
	discard
)

type stdioType int

func (t stdioType) String() string {
	switch t {
	case stdout:
		return "STDOUT"
	case stdin:
		return "STDIN"
	case stderr:
		return "STDERR"
	case discard:
		return "DISCARD"
	}
	return "INVALID"
}

// IO wraps the three container's ios of stdout, stderr, stdin.
type IO struct {
	Stdout *ContainerIO
	Stderr *ContainerIO
	Stdin  *ContainerIO

	// For IO backend like http, we need to mux stdout & stderr
	// if terminal is disabled.
	// But for other IO backend, it is not necessary.
	// So we should make it configurable.
	MuxDisabled bool
}

// NewIO creates the container's ios of stdout, stderr, stdin.
func NewIO(opt *Option) *IO {
	backends := createBackend(opt)

	i := &IO{
		Stdout:      create(opt, stdout, backends),
		Stderr:      create(opt, stderr, backends),
		MuxDisabled: opt.muxDisabled,
	}

	if opt.stdin {
		i.Stdin = create(opt, stdin, backends)
	}

	return i
}

// AddBackend adds more backends to container's stdio.
func (io *IO) AddBackend(opt *Option) {
	backends := createBackend(opt)

	for t, s := range map[stdioType]*ContainerIO{
		stdout: io.Stdout,
		stderr: io.Stderr,
	} {
		s.add(opt, t, backends)
	}

	if opt.stdin && io.Stdin != nil {
		io.Stdin.add(opt, stdin, backends)
	}
}

// Close closes the container's io.
func (io *IO) Close() error {
	io.Stderr.Close()
	io.Stdout.Close()
	if io.Stdin != nil {
		io.Stdin.Close()
	}
	return nil
}

// ContainerIO used to control the container's stdio.
type ContainerIO struct {
	Option
	backends []containerBackend
	total    int64
	typ      stdioType
	closed   bool
	// The stdin of all backends should put into ring first.
	ring *ringbuffer.RingBuffer
}

func (cio *ContainerIO) add(opt *Option, typ stdioType, backends map[string]containerBackend) {
	if typ == stdin {
		for _, b := range backends {
			if b.backend.Name() == opt.stdinBackend {
				cio.backends = append(cio.backends, b)
				go func(b containerBackend) {
					cio.converge(b.backend.Name(), opt.id, b.backend.In())
					b.backend.Close()
				}(b)
				break
			}
		}
	} else {
		for _, b := range backends {
			cio.backends = append(cio.backends, b)
		}
	}
}

func create(opt *Option, typ stdioType, backends map[string]containerBackend) *ContainerIO {
	io := &ContainerIO{
		total:  0,
		typ:    typ,
		closed: false,
		Option: *opt,
	}

	if typ == stdin {
		io.ring = ringbuffer.New(-1)
		for _, b := range backends {
			if b.backend.Name() == opt.stdinBackend {
				io.backends = append(io.backends, b)
				go func(b containerBackend) {
					// For backend with stdin, close it if stdin finished.
					io.converge(b.backend.Name(), opt.id, b.backend.In())
					b.backend.Close()
				}(b)
				break
			}
		}
	} else {
		for _, b := range backends {
			io.backends = append(io.backends, b)
		}
	}

	return io
}

func createBackend(opt *Option) map[string]containerBackend {
	backends := make(map[string]containerBackend)

	for _, create := range backendFactorys {
		backend := create()
		if _, ok := opt.backends[backend.Name()]; !ok {
			continue
		}

		if err := backend.Init(opt); err != nil {
			// FIXME skip the backend.
			logrus.Errorf("failed to initialize backend: %s, id: %s, %v", backend.Name(), opt.id, err)
			continue
		}

		backends[backend.Name()] = containerBackend{
			backend: backend,
			outRing: ringbuffer.New(-1),
			errRing: ringbuffer.New(-1),
		}
	}

	// start to subscribe stdout and stderr ring buffer.
	for _, b := range backends {

		// the goroutine don't exit forever.
		go func(b containerBackend) {
			subscribe(b.backend.Name(), opt.id, b.outRing, b.backend.Out())
		}(b)
		go func(b containerBackend) {
			subscribe(b.backend.Name(), opt.id, b.errRing, b.backend.Err())
		}(b)
	}

	return backends
}

// OpenStdin returns open container's stdin or not.
func (cio *ContainerIO) OpenStdin() bool {
	if cio.typ != stdin {
		return false
	}
	if cio.closed {
		return false
	}
	return len(cio.backends) != 0
}

// Read implements the standard Read interface.
func (cio *ContainerIO) Read(p []byte) (int, error) {
	if cio.typ != stdin {
		return 0, fmt.Errorf("invalid container io type: %s, id: %s", cio.typ, cio.id)
	}
	if cio.closed {
		return 0, fmt.Errorf("container io is closed")
	}

	value, _ := cio.ring.Pop()
	data, ok := value.([]byte)
	if !ok {
		return 0, nil
	}
	n := copy(p, data)

	return n, nil
}

// Write implements the standard Write interface.
func (cio *ContainerIO) Write(data []byte) (int, error) {
	if cio.typ == stdin {
		return 0, fmt.Errorf("invalid container io type: %s, id: %s", cio.typ, cio.id)
	}
	if cio.closed {
		return 0, fmt.Errorf("container io is closed")
	}

	if cio.typ == discard {
		return len(data), nil
	}

	// FIXME(fuwei): In case that the data slice is reused by the writer,
	// we should copy the data before we push it into the ringbuffer.
	// The previous data shares the same address with the coming data.
	// If we don't copy the data and the previous data isn't consumed by
	// ringbuf pop action, the incoming data will override the previous data
	// in the ringbuf.
	//
	// However, copy data maybe impact the performance. We need to reconsider
	// other better way to handle the IO.
	copyData := make([]byte, len(data))
	copy(copyData, data)

	switch cio.typ {
	case stdout:
		for _, b := range cio.backends {
			cover, err := b.outRing.Push(copyData)
			// skip if it is closed ringbuffer
			if err != nil {
				continue
			}

			if cover {
				logrus.Warnf("cover stdout data, backend: %s, id: %s", b.backend.Name(), cio.id)
			}
		}
	case stderr:
		for _, b := range cio.backends {
			cover, err := b.errRing.Push(copyData)
			// skip if it is closed ringbuffer
			if err != nil {
				continue
			}

			if cover {
				logrus.Warnf("cover stderr data, backend: %s, id: %s", b.backend.Name(), cio.id)
			}
		}
	}

	return len(data), nil
}

// Close implements the standard Close interface.
func (cio *ContainerIO) Close() error {
	// FIXME(fuwei): stdin should be treated like stdout, stderr.
	if cio.typ == stdin && cio.ring != nil {
		// NOTE: let converge goroutine quit
		cio.ring.Close()
	}

	for _, b := range cio.backends {
		// we need to close ringbuf before close backend, because close ring will flush
		// the remain data into backend.
		name := b.backend.Name()

		b.outRing.Close()
		b.errRing.Close()
		if err := b.drainRingBuffer(); err != nil {
			logrus.Warnf("failed to drain ringbuffer for backend: %s, id: %s", name, cio.id)
		}

		b.backend.Close()
		logrus.Infof("close containerio backend: %s, id: %s", name, cio.id)
	}

	cio.closed = true
	return nil
}

// FIXME(fuwei): just one ringbuffer for one backend
type containerBackend struct {
	backend Backend
	outRing *ringbuffer.RingBuffer
	errRing *ringbuffer.RingBuffer
}

func (cb *containerBackend) drainRingBuffer() error {
	for _, item := range []struct {
		data []interface{}
		w    io.Writer
	}{
		{data: cb.outRing.Drain(), w: cb.backend.Out()},
		{data: cb.errRing.Drain(), w: cb.backend.Err()},
	} {
		for _, value := range item.data {
			if b, ok := value.([]byte); ok {
				if _, err := item.w.Write(b); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// subscribe be called in a groutine.
func subscribe(name, id string, ring *ringbuffer.RingBuffer, out io.Writer) {
	logrus.Infof("start to subscribe io, backend: %s, id: %s", name, id)

	for {
		value, err := ring.Pop()
		// break loop if the ringbuffer has been closed
		if err != nil {
			break
		}

		if b, ok := value.([]byte); ok {
			if _, err := out.Write(b); err != nil {
				logrus.Errorf("failed to write containerio backend: %s, id: %s, %v", name, id, err)
			}
		}
	}
	logrus.Infof("finished to subscribe io, backend: %s, id: %s", name, id)
}

// converge be called in a goroutine.
func (cio *ContainerIO) converge(name, id string, in io.Reader) {
	// TODO: we should implement this function more elegant and robust.
	logrus.Infof("start to converge io, backend: %s, id: %s", name, id)

	data := make([]byte, 128)
	for {
		n, err := in.Read(data)
		if err != nil {
			logrus.Errorf("failed to read from backend: %s, id: %s, %v", name, id, err)
			break
		}

		// FIXME(fuwei): In case that the data slice is reused by the writer,
		// we should copy the data before we push it into the ringbuffer.
		// The previous data shares the same address with the coming data.
		// If we don't copy the data and the previous data isn't consumed by
		// ringbuf pop action, the incoming data will override the previous data
		// in the ringbuf.
		copyData := make([]byte, n)
		copy(copyData, data[:n])

		cover, err := cio.ring.Push(copyData)
		if err != nil {
			break
		}

		if cover {
			logrus.Warnf("cover data, backend: %s, id: %s", name, id)
		}
	}
	logrus.Infof("finished to converge io, backend: %s, id: %s", name, id)
}
