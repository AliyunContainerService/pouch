package containerio

import (
	"fmt"
	"io"

	"github.com/alibaba/pouch/pkg/ringbuff"

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
		Stdin:       create(opt, stdin, backends),
		MuxDisabled: opt.muxDisabled,
	}

	/*
		NOTE(tanghuamin): for compatible docker 1.12
		if opt.stdin {
			i.Stdin = create(opt, stdin, backends)
		}
	*/

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
	ring *ringbuff.RingBuff
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
		io.ring = ringbuff.New(10)
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
			outRing: ringbuff.New(10),
			errRing: ringbuff.New(10),
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

	switch cio.typ {
	case stdout:
		for _, b := range cio.backends {
			if cover := b.outRing.Push(data); cover {
				logrus.Warnf("cover data, backend: %s, id: %s", b.backend.Name(), cio.id)
			}
		}
	case stderr:
		for _, b := range cio.backends {
			if cover := b.errRing.Push(data); cover {
				logrus.Warnf("cover data, backend: %s, id: %s", b.backend.Name(), cio.id)
			}
		}
	}

	return len(data), nil
}

// Close implements the standard Close interface.
func (cio *ContainerIO) Close() error {
	for _, b := range cio.backends {
		// we need to close ringbuf before close backend, because close ring will flush
		// the remain data into backend.
		name := b.backend.Name()
		b.outRing.Close()
		b.errRing.Close()
		b.backend.Close()

		logrus.Infof("close containerio backend: %s, id: %s", name, cio.id)
	}

	cio.closed = true
	return nil
}

type containerBackend struct {
	backend Backend
	outRing *ringbuff.RingBuff
	errRing *ringbuff.RingBuff
}

// subscribe be called in a groutine.
func subscribe(name, id string, ring *ringbuff.RingBuff, out io.Writer) {
	logrus.Infof("start to subscribe io, backend: %s, id: %s", name, id)

	for {
		value, closed := ring.Pop() // will block, if no element.

		if b, ok := value.([]byte); ok {
			if _, err := out.Write(b); err != nil {
				logrus.Errorf("failed to write containerio backend: %s, id: %s, %v", name, id, err)
			}
		}

		if value == nil && closed {
			break
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
		cover := cio.ring.Push(data[:n])
		if cover {
			logrus.Warnf("cover data, backend: %s, id: %s", name, id)
		}
	}

	logrus.Infof("finished to converge io, backend: %s, id: %s", name, id)
}
