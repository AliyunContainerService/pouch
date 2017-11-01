package config

import (
	"github.com/alibaba/pouch/volume/types"
)

// Config refers to daemon's whole configurations.
type Config struct {
	//Volume config
	types.Config

	// Server listening address.
	Listen []string

	// Debug refers to the log mode.
	Debug bool

	// ContainerdAddr refers to the unix socket path of containerd.
	ContainerdAddr string

	// DefaultRegistry is daemon's default registry which is to pull/push/search images.
	DefaultRegistry string

	// Home directory.
	HomeDir string

	// CotnainerdPath is the absolute path of containerd binary,
	// /usr/local/bin is the default.
	ContainerdPath string

	// Containerd's config file.
	ContainerdConfig string
}
