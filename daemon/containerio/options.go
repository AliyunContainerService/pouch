package containerio

import (
	"io"
	"net/http"
	"os"

	"github.com/alibaba/pouch/cri/stream/remotecommand"
	"github.com/alibaba/pouch/daemon/logger"
)

// Option is used to pass some data into ContainerIO.
//
// FIXME(fuwei): use logger.Info to separate options and backends.
type Option struct {
	info logger.Info

	id            string
	stdin         bool
	muxDisabled   bool
	backends      map[string]struct{}
	hijack        http.Hijacker
	hijackUpgrade bool
	stdinBackend  string
	pipe          *io.PipeWriter
	streams       *remotecommand.Streams
	criLogFile    *os.File
}

// NewOption creates the Option instance.
func NewOption(opts ...func(*Option)) *Option {
	opt := &Option{}

	for _, o := range opts {
		o(opt)
	}

	return opt
}

// WithID specified the container's id.
func WithID(id string) func(*Option) {
	return func(opt *Option) {
		opt.id = id
	}
}

// WithLoggerInfo specified the container's logger information.
func WithLoggerInfo(info logger.Info) func(*Option) {
	return func(opt *Option) {
		opt.info = info
	}
}

// WithStdin specified whether open the container's stdin.
func WithStdin(stdin bool) func(*Option) {
	return func(opt *Option) {
		opt.stdin = stdin
	}
}

// WithMuxDisabled specified whether mux stdout & stderr of container IO.
func WithMuxDisabled(muxDisabled bool) func(*Option) {
	return func(opt *Option) {
		opt.muxDisabled = muxDisabled
	}
}

// WithDiscard specified the discard backend.
func WithDiscard() func(*Option) {
	return func(opt *Option) {
		if opt.backends == nil {
			opt.backends = make(map[string]struct{})
		}
		opt.backends["discard"] = struct{}{}
	}
}

// WithJSONFile specified the jsonfile backend.
func WithJSONFile() func(*Option) {
	return func(opt *Option) {
		if opt.backends == nil {
			opt.backends = make(map[string]struct{})
		}
		opt.backends["jsonfile"] = struct{}{}
	}
}

// WithHijack specified the hijack backend.
func WithHijack(hi http.Hijacker, upgrade bool) func(*Option) {
	return func(opt *Option) {
		if opt.backends == nil {
			opt.backends = make(map[string]struct{})
		}
		opt.backends["hijack"] = struct{}{}
		opt.hijack = hi
		opt.hijackUpgrade = upgrade
	}
}

// WithStdinHijack sepcified the stdin with hijack.
func WithStdinHijack() func(*Option) {
	return func(opt *Option) {
		opt.stdinBackend = "hijack"
	}
}

// WithPipe specified the pipe backend.
func WithPipe(pipe *io.PipeWriter) func(*Option) {
	return func(opt *Option) {
		if opt.backends == nil {
			opt.backends = make(map[string]struct{})
		}
		opt.backends["pipe"] = struct{}{}
		opt.pipe = pipe
	}
}

// WithStreams specified the stream backend.
func WithStreams(streams *remotecommand.Streams) func(*Option) {
	return func(opt *Option) {
		if opt.backends == nil {
			opt.backends = make(map[string]struct{})
		}
		opt.backends["streams"] = struct{}{}
		opt.streams = streams
	}
}

// WithStdinStream specified the stdin with stream.
func WithStdinStream() func(*Option) {
	return func(opt *Option) {
		opt.stdinBackend = "streams"
	}
}

// WithCriLogFile specified the cri log file backend.
func WithCriLogFile(criLogFile *os.File) func(*Option) {
	return func(opt *Option) {
		if opt.backends == nil {
			opt.backends = make(map[string]struct{})
		}
		opt.backends["cri-log-file"] = struct{}{}
		opt.criLogFile = criLogFile
	}
}
