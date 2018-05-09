package remote

import (
	"github.com/alibaba/pouch/storage/volume"
	"github.com/alibaba/pouch/storage/volume/driver"
	"github.com/alibaba/pouch/storage/volume/types"
)

// DriverWrapper represents a volume driver.
type DriverWrapper struct {
	driverName string
	proxy      *remoteDriverProxy
}

// Name returns remote volume driver's name.
func (r *DriverWrapper) Name(ctx driver.Context) string {
	return r.driverName
}

// StoreMode returns remote volume driver's store mode.
func (r *DriverWrapper) StoreMode(ctx driver.Context) driver.VolumeStoreMode {
	return driver.RemoteStore | driver.UseLocalMetaStore
}

// Create a remote volume.
func (r *DriverWrapper) Create(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("driver wrapper %s creates volume: %s", r.Name(ctx), v.Name)

	options := volume.ExtractOptionsFromVolume(v)

	ctx.Log.Debugf("driver wrapper gets options: %v", options)

	return r.proxy.Create(v.Name, options)
}

// Remove a remote volume.
func (r *DriverWrapper) Remove(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("driver wrapper %s removes volume: %s", r.Name(ctx), v.Name)

	return r.proxy.Remove(v.Name)
}

// Path returns remote volume mount path.
func (r *DriverWrapper) Path(ctx driver.Context, v *types.Volume) (string, error) {
	ctx.Log.Debugf("driver wrapper %s get volume %s mount path", r.Name(ctx), v.Name)

	// Get the mount path from remote plugin
	mountPath, err := r.proxy.Path(v.Name)
	if err != nil {
		return "", err
	}
	return mountPath, nil
}

// Options returns remote volume options.
func (r *DriverWrapper) Options() map[string]types.Option {
	return map[string]types.Option{}
}

// Attach a remote volume.
func (r *DriverWrapper) Attach(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("driver wrapper %s attach volume: %s", r.Name(ctx), v.Name)

	_, err := r.proxy.Mount(v.Name, v.UID)
	if err != nil {
		return err
	}

	return nil
}

// Detach a remote volume.
func (r *DriverWrapper) Detach(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("driver wrapper %s detach volume: %s", r.Name(ctx), v.Name)

	return r.proxy.Unmount(v.Name, v.UID)
}
