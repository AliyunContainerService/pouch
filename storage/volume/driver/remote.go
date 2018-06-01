package driver

import (
	"github.com/alibaba/pouch/plugins"
	"github.com/alibaba/pouch/storage/volume/types"
)

// remoteDriverWrapper represents a volume driver.
type remoteDriverWrapper struct {
	driverName string
	proxy      *remoteDriverProxy
}

// NewRemoteDriverWrapper returns a remote driver
func NewRemoteDriverWrapper(name string, plugin *plugins.Plugin) Driver {
	return &remoteDriverWrapper{
		driverName: name,
		proxy: &remoteDriverProxy{
			Name:   name,
			client: plugin.Client(),
		},
	}
}

// Name returns the volume driver's name.
func (r *remoteDriverWrapper) Name(ctx Context) string {
	return r.driverName
}

// StoreMode returns the volume driver's store mode.
func (r *remoteDriverWrapper) StoreMode(ctx Context) VolumeStoreMode {
	return RemoteStore | UseLocalMetaStore
}

// Create a remote volume.
func (r *remoteDriverWrapper) Create(ctx Context, id types.VolumeID) (*types.Volume, error) {
	ctx.Log.Debugf("driver wrapper [%s] creates volume: %s", r.Name(ctx), id.Name)

	ctx.Log.Debugf("driver wrapper gets options: %v", id.Options)

	if err := r.proxy.Create(id.Name, id.Options); err != nil {
		return nil, err
	}

	mountPath, err := r.proxy.Path(id.Name)
	if err != nil {
		mountPath = ""
	}

	return types.NewVolumeFromID(mountPath, "", id), nil
}

// Remove a remote volume.
func (r *remoteDriverWrapper) Remove(ctx Context, v *types.Volume) error {
	ctx.Log.Debugf("driver wrapper [%s] removes volume: %s", r.Name(ctx), v.Name)

	return r.proxy.Remove(v.Name)
}

// Get a volume from remote driver.
func (r *remoteDriverWrapper) Get(ctx Context, name string) (*types.Volume, error) {
	ctx.Log.Debugf("driver wrapper [%s] gets volume: %s", r.Name(ctx), name)

	rv, err := r.proxy.Get(name)
	if err != nil {
		return nil, err
	}

	id := types.NewVolumeID(name, r.Name(ctx))

	return types.NewVolumeFromID(rv.Mountpoint, "", id), nil
}

// List all volumes from remote driver.
func (r *remoteDriverWrapper) List(ctx Context) ([]*types.Volume, error) {
	ctx.Log.Debugf("driver wrapper [%s] list all volumes", r.Name(ctx))

	rvList, err := r.proxy.List()
	if err != nil {
		return nil, err
	}

	var vList []*types.Volume

	for _, rv := range rvList {
		id := types.NewVolumeID(rv.Name, r.Name(ctx))
		volume := types.NewVolumeFromID(rv.Mountpoint, "", id)
		vList = append(vList, volume)
	}

	return vList, nil

}

// Path returns remote volume mount path.
func (r *remoteDriverWrapper) Path(ctx Context, v *types.Volume) (string, error) {
	ctx.Log.Debugf("driver wrapper [%s] gets volume [%s] mount path", r.Name(ctx), v.Name)

	// Get the mount path from remote plugin
	mountPath, err := r.proxy.Path(v.Name)
	if err != nil {
		return "", err
	}
	return mountPath, nil
}

// Options returns remote volume options.
func (r *remoteDriverWrapper) Options() map[string]types.Option {
	return map[string]types.Option{}
}

// Attach a remote volume.
func (r *remoteDriverWrapper) Attach(ctx Context, v *types.Volume) error {
	ctx.Log.Debugf("driver wrapper [%s] attach volume: %s", r.Name(ctx), v.Name)

	_, err := r.proxy.Mount(v.Name, v.UID)
	if err != nil {
		return err
	}

	return nil
}

// Detach a remote volume.
func (r *remoteDriverWrapper) Detach(ctx Context, v *types.Volume) error {
	ctx.Log.Debugf("driver wrapper [%s] detach volume: %s", r.Name(ctx), v.Name)

	return r.proxy.Unmount(v.Name, v.UID)
}
