package containerio

import (
	"net/http"
)

// Option is used to pass some data into ContainerIO.
type Option struct {
	id            string
	rootDir       string
	backends      map[string]struct{}
	hijack        http.Hijacker
	hijackUpgrade bool
	stdinBackend  string
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

// WithRootDir specified the container's root dir.
func WithRootDir(dir string) func(*Option) {
	return func(opt *Option) {
		opt.rootDir = dir
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

// WithRawFile specified the raw-file backend.
func WithRawFile() func(*Option) {
	return func(opt *Option) {
		if opt.backends == nil {
			opt.backends = make(map[string]struct{})
		}
		opt.backends["raw-file"] = struct{}{}
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
