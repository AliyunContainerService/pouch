package supervisord

import "github.com/BurntSushi/toml"

// Config provides containerd configuration data for the server
//
// NOTE: The config is copied from containerd@v1.0.3/server, because the server
// package depends on a lot of packages which pouch doesn't need.
type Config struct {
	// Root is the path to a directory where containerd will store persistent data
	Root string `toml:"root"`
	// State is the path to a directory where containerd will store transient data
	State string `toml:"state"`
	// GRPC configuration settings
	GRPC GRPCConfig `toml:"grpc"`
	// Debug and profiling settings
	Debug Debug `toml:"debug"`
	// Metrics and monitoring settings
	Metrics MetricsConfig `toml:"metrics"`
	// Plugins provides plugin specific configuration for the initialization of a plugin
	Plugins map[string]toml.Primitive `toml:"plugins"`
	// OOMScore adjust the containerd's oom score
	OOMScore int `toml:"oom_score"`
	// Cgroup specifies cgroup information for the containerd daemon process
	Cgroup CgroupConfig `toml:"cgroup"`

	md toml.MetaData
}

// GRPCConfig provides GRPC configuration for the socket
type GRPCConfig struct {
	Address string `toml:"address"`
	UID     int    `toml:"uid"`
	GID     int    `toml:"gid"`
}

// Debug provides debug configuration
type Debug struct {
	Address string `toml:"address"`
	UID     int    `toml:"uid"`
	GID     int    `toml:"gid"`
	Level   string `toml:"level"`
}

// MetricsConfig provides metrics configuration
type MetricsConfig struct {
	Address string `toml:"address"`
}

// CgroupConfig provides cgroup configuration
type CgroupConfig struct {
	Path string `toml:"path"`
}
