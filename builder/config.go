package builder

import "github.com/moby/buildkit/util/appdefaults"

// Config is used to set up builder.
type Config struct {
	Debug bool

	Root string

	GRPC struct {
		Address  string
		UID, GID int
	}
	ContainerdWorker struct {
		Address     string
		Namespace   string
		Snapshotter string
	}
}

// setDefaultConfig sets default value if missing.
func setDefaultConfig(cfg *Config) {
	if cfg.Root == "" {
		cfg.Root = appdefaults.Root
	}

	if cfg.GRPC.Address == "" {
		cfg.GRPC.Address = appdefaults.Address
	}
}
