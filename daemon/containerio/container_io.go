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
}

// NewIO creates the container's ios of stdout, stderr, stdin.
func NewIO(opt *Option) *IO {
	backends := createBackend(opt)

	return &IO{
		Stdout: create(opt, stdout, backends),
		Stderr: create(opt, stderr, backends),
		Stdin:  create(opt, stdin, backends),
	}
}

// ContainerIO used to control the container's stdio.
type ContainerIO struct {
	Option
	backends map[string]containerBackend
	total    int64
	typ      stdioType
	closed   bool
}

func create(opt *Option, typ stdioType, backends map[string]containerBackend) *ContainerIO {
	io := &ContainerIO{
		backends: backends,
		total:    0,
		typ:      typ,
		closed:   false,
		Option:   *opt,
	}

	if typ == stdin {
		io.backends = make(map[string]containerBackend)

		for _, b := range backends {
			if b.backend.Name() == opt.stdinBackend {
				io.backends[opt.stdinBackend] = b
				break
			}
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
			ring:    ringbuff.New(10),
		}
	}

	// start to subscribe ring buffer.
	for _, b := range backends {

		// the goroutine don't exit forever.
		go func(b containerBackend) {
			subscribe(b.backend.Name(), opt.id, b.ring, b.backend.Out())
		}(b)
	}

	return backends
}

// OpenStdin returns open container's stdin or not.
func (io *ContainerIO) OpenStdin() bool {
	if io.typ != stdin {
		return false
	}
	if io.closed {
		return false
	}
	return len(io.backends) != 0
}

// Read implements the standard Read interface.
func (io *ContainerIO) Read(p []byte) (int, error) {
	if io.typ != stdin {
		return 0, fmt.Errorf("invalid container io type: %s, id: %s", io.typ, io.id)
	}
	if io.closed {
		return 0, fmt.Errorf("container io is closed")
	}

	if len(io.backends) == 0 {
		block := make(chan struct{})
		<-block
	}

	backend := io.backends[io.stdinBackend]

	return backend.backend.In().Read(p)
}

// Write implements the standard Write interface.
func (io *ContainerIO) Write(data []byte) (int, error) {
	if io.typ == stdin {
		return 0, fmt.Errorf("invalid container io type: %s, id: %s", io.typ, io.id)
	}
	if io.closed {
		return 0, fmt.Errorf("container io is closed")
	}

	if io.typ == discard {
		return len(data), nil
	}

	for _, b := range io.backends {
		if cover := b.ring.Push(data); cover {
			logrus.Warnf("cover data, backend: %s, id: %s", b.backend.Name(), io.id)
		}
	}

	return len(data), nil
}

// Close implements the standard Close interface.
func (io *ContainerIO) Close() error {
	for name, b := range io.backends {
		b.backend.Close()
		b.ring.Close()

		logrus.Infof("close containerio backend: %s, id: %s", name, io.id)
	}

	io.closed = true
	return nil
}

type containerBackend struct {
	backend Backend
	ring    *ringbuff.RingBuff
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
