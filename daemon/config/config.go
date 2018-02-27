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
	Listen []string `json:"listen,omitempty"`

	// ListenCRI is the listening address which serves CRI.
	ListenCRI string `json:"listen-cri,omitempty"`

	// Debug refers to the log mode.
	Debug bool `json:"debug,omitempty"`

	// ContainerdAddr refers to the unix socket path of containerd.
	ContainerdAddr string `json:"containerd,omitempty"`

	// DefaultRegistry is daemon's default registry which is to pull/push/search images.
	DefaultRegistry string

	// Home directory.
	HomeDir string `json:"home-dir,omitempty"`

	// ContainerdPath is the absolute path of containerd binary,
	// /usr/local/bin is the default.
	ContainerdPath string `json:"containerd-path"`

	// TLS configuration
	TLS utils.TLSConfig

	// Default OCI Runtime
	DefaultRuntime string `json:"default-runtime,omitempty"`

	// Enable lxcfs
	IsLxcfsEnabled bool `json:"enable-lxcfs,omitempty"`

	// LxcfsBinPath is the absolute path of lxcfs binary
	LxcfsBinPath string `json:"lxcfs,omitempty"`

	// LxcfsHome is the absolute path of lxcfs
	LxcfsHome string

	// ImageProxy is a http proxy to pull image
	ImageProxy string `json:"image-proxy,omitempty"`

	// QuotaDriver is used to set the driver of Quota
	QuotaDriver string

	// Configuration file of pouchd
	ConfigFile string `json:"config-file,omitempty"`

	// CgroupParent is to set parent cgroup for all containers
	CgroupParent string `json:"cgroup-parent,omitempty"`
}
