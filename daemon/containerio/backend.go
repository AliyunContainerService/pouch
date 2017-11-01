package containerio

import (
	"io"
)

var backendFactorys []func() Backend

// Backend defines the real output/input of the container's stdio.
type Backend interface {
	// Name defines the backend's name.
	Name() string

	// Init initializes the backend io.
	Init(opt *Option) error

	// Out returns the stdout/stderr.
	Out() io.Writer

	// In returns the stdin.
	In() io.Reader

	// Close closes the io.
	Close() error
}

// Register adds a backend.
func Register(create func() Backend) {
	backendFactorys = append(backendFactorys, create)
}
