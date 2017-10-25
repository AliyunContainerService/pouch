package driver

import (
	"github.com/alibaba/pouch/volume/types"
)

// Driver represents volume driver base operation interface.
type Driver interface {
	// Name returns backend driver's name.
	Name(Context) string

	// StoreMode defines backend driver's store model.
	StoreMode(Context) VolumeStoreMode

	// Create a volume.
	Create(Context, *types.Volume, *types.Storage) error

	// Remove a volume.
	Remove(Context, *types.Volume, *types.Storage) error

	// Path returns volume's path.
	Path(Context, *types.Volume) (string, error)
}

// Opt represents volume driver option interface.
type Opt interface {
	// Options return module customize volume options.
	Options() map[string]types.Option
}

// AttachDetach represents volume attach/detach interface.
type AttachDetach interface {
	// Attach a Volume to host, enable the volume.
	Attach(Context, *types.Volume, *types.Storage) error

	// Detach a volume with host, disable the volume.
	Detach(Context, *types.Volume, *types.Storage) error
}

// Formator represents volume format interface.
type Formator interface {
	// Format a volume.
	Format(Context, *types.Volume, *types.Storage) error
}

// GatewayDriver represents storage gateway interface.
type GatewayDriver interface {
	// Report storage cluster status.
	Report(Context) ([]*types.Storage, error)
}
