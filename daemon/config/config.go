package config

import (
	"github.com/alibaba/pouch/cri"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/volume"
)

// Config refers to daemon's whole configurations.
type Config struct {
	//Volume config
	VolumeConfig volume.Config

	// Network config
	NetworkConfg network.Config

	// CRI config.
	CriConfig cri.Config

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

	// ContainerdPath is the absolute path of containerd binary,
	// /usr/local/bin is the default.
	ContainerdPath string

	// TLS configuration
	TLS utils.TLSConfig

	// Default OCI Runtime
	DefaultRuntime string

	// Enable lxcfs
	IsLxcfsEnabled bool

	// LxcfsBinPath is the absolute path of lxcfs binary
	LxcfsBinPath string

	// LxcfsHome is the absolute path of lxcfs
	LxcfsHome string

	// ImageProxy is a http proxy to pull image
	ImageProxy string `json:"image-proxy,omitempty"`

	// QuotaDriver is used to set the driver of Quota
	QuotaDriver string
}
