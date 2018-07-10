package driver

import (
	volerr "github.com/alibaba/pouch/storage/volume/error"
	"path"

	"github.com/alibaba/pouch/storage/volume/types"
)

// FakeDriver is a fake volume driver.
type FakeDriver struct {
	name string
}

// NewFakeDriver returns a fake voluem driver.
func NewFakeDriver(name string) Driver {
	return &FakeDriver{
		name: name,
	}
}

// Name returns the fake driver's name.
func (f *FakeDriver) Name(ctx Context) string {
	return f.name
}

// StoreMode returns the fake driver's store model.
func (f *FakeDriver) StoreMode(ctx Context) VolumeStoreMode {
	return UseLocalMetaStore | LocalStore
}

// Create a fake volume
func (f *FakeDriver) Create(ctx Context, id types.VolumeID) (*types.Volume, error) {
	// generate the mountPath
	mountPath := path.Join("/fake", id.Name)

	return types.NewVolumeFromID(mountPath, "", id), nil
}

// Remove a fake volume
func (f *FakeDriver) Remove(ctx Context, volume *types.Volume) error {
	return nil
}

// Path returns fake volume's path.
func (f *FakeDriver) Path(ctx Context, volume *types.Volume) (string, error) {
	return path.Join("/fake", volume.Name), nil
}

// Get a volume from fake driver.
func (f *FakeDriver) Get(ctx Context, name string) (*types.Volume, error) {

	if name == "test1" {
		id := types.NewVolumeID(name, f.Name(ctx))
		mountPath := path.Join("/fake", id.Name)
		return types.NewVolumeFromID(mountPath, "", id), nil
	}
	return nil, volerr.ErrVolumeNotFound

}

// List all volumes from fake driver.
func (f *FakeDriver) List(ctx Context) ([]*types.Volume, error) {
	var vList []*types.Volume

	id := types.NewVolumeID("fake2", f.Name(ctx))
	mountPath := path.Join("/fake", id.Name)
	volume := types.NewVolumeFromID(mountPath, "", id)
	vList = append(vList, volume)

	return vList, nil
}

// Attach a fake volume.
func (f *FakeDriver) Attach(ctx Context, v *types.Volume) error {
	return nil
}

// Detach a fake volume.
func (f *FakeDriver) Detach(ctx Context, v *types.Volume) error {
	return nil
}
