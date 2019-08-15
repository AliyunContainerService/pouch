package driver

import (
	"context"

	"github.com/alibaba/pouch/storage/volume/types"
)

// Driver represents volume driver base operation interface.
type Driver interface {
	// Name returns backend driver's name.
	Name(context.Context) string

	// StoreMode defines backend driver's store model.
	StoreMode(context.Context) VolumeStoreMode

	// Create a volume.
	Create(context.Context, types.VolumeContext) (*types.Volume, error)

	// Remove a volume.
	Remove(context.Context, *types.Volume) error

	// Path returns volume's path.
	Path(context.Context, *types.Volume) (string, error)
}

// Opt represents volume driver option interface.
type Opt interface {
	// Options return module customize volume options.
	Options() map[string]types.Option
}

// Conf represents pass volume config to volume driver.
type Conf interface {
	// Config is used to pass the daemon volume config into driver.
	Config(context.Context, map[string]interface{}) error
}

// AttachDetach represents volume attach/detach interface.
type AttachDetach interface {
	// Attach a Volume to host, enable the volume.
	Attach(context.Context, *types.Volume) error

	// Detach a volume with host, disable the volume.
	Detach(context.Context, *types.Volume) error
}

// Formator represents volume format interface.
type Formator interface {
	// Format a volume.
	Format(context.Context, *types.Volume) error
}

// Getter represents volume get interface.
type Getter interface {
	// Get a volume from driver
	Get(context.Context, string) (*types.Volume, error)
}

// Lister represents volume list interface
type Lister interface {
	// List a volume from driver
	List(context.Context) ([]*types.Volume, error)
}
