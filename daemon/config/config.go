package config

// Config refers to daemon's whole configurations.
type Config struct {
	//
	Listen []string
	// Debug refers to the log mode.
	Debug bool
	// ContainerdAddr refers to the unix socket path of containerd.
	ContainerdAddr string
	// DefaultRegistry is daemon's default registry which is to pull/push/search images.
	DefaultRegistry string
}
