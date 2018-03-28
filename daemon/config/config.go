package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/cri"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/volume"
)

// Config refers to daemon's whole configurations.
type Config struct {
	sync.Mutex

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

	// DefaultRegistryNS is daemon's default registry namespace used in pull/push/search images.
	DefaultRegistryNS string

	// Home directory.
	HomeDir string `json:"home-dir,omitempty"`

	// ContainerdPath is the absolute path of containerd binary,
	// /usr/local/bin is the default.
	ContainerdPath string `json:"containerd-path"`

	// TLS configuration
	TLS client.TLSConfig

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

	// PluginPath is set the path where plugin so file put
	PluginPath string `json:"plugin"`

	// Labels is the metadata of daemon
	Labels []string `json:"labels,omitempty"`
}

// Validate validates the user input config.
func (cfg *Config) Validate() error {
	// deduplicated elements in slice if there is any.
	cfg.Listen = utils.DeDuplicate(cfg.Listen)
	cfg.Labels = utils.DeDuplicate(cfg.Labels)

	for _, label := range cfg.Labels {
		data := strings.SplitN(label, "=", 2)
		if len(data) != 2 {
			return fmt.Errorf("daemon label %s must be in format of key=value", label)
		}
		if len(data[0]) == 0 || len(data[1]) == 0 {
			return fmt.Errorf("key and value in daemon label %s cannot be empty", label)
		}
	}

	// TODO: add config validation

	return nil
}
