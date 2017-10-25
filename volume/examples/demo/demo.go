package demo

import (
	"path"

	"github.com/alibaba/pouch/volume/driver"
	"github.com/alibaba/pouch/volume/types"
)

func init() {
	if err := driver.Register(&Demo{}); err != nil {
		panic(err)
	}
}

// Demo represents demo volume driver.
type Demo struct {
}

// Name returns volume driver's name.
func (d *Demo) Name(ctx driver.Context) string {
	return "demo"
}

// StoreMode returns demo driver's store mode.
func (d *Demo) StoreMode(ctx driver.Context) driver.VolumeStoreMode {
	return driver.RemoteStore
}

// Create a demo volume.
func (d *Demo) Create(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Infof("Demo create volume: %s", v.Name)
	return nil
}

// Remove a demo volume.
func (d *Demo) Remove(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Infof("Demo Remove volume: %s", v.Name)
	return nil
}

// Path returns demo volume path.
func (d *Demo) Path(ctx driver.Context, v *types.Volume) (string, error) {
	ctx.Log.Infof("Demo volume path: %s", v.Name)
	return path.Join("/mnt", d.Name(ctx), v.Name), nil
}
