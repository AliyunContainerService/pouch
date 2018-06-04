// +build linux

package local

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/alibaba/pouch/pkg/bytefmt"
	"github.com/alibaba/pouch/storage/quota"
	"github.com/alibaba/pouch/storage/volume/driver"
	"github.com/alibaba/pouch/storage/volume/types"
)

var (
	dataDir = "/var/lib/pouch/volume"
)

func init() {
	if err := driver.Register(&Local{}); err != nil {
		panic(err)
	}
}

// Local represents local volume driver.
type Local struct {
}

// Name returns local volume driver's name.
func (p *Local) Name(ctx driver.Context) string {
	return "local"
}

// StoreMode returns local volume driver's store mode.
func (p *Local) StoreMode(ctx driver.Context) driver.VolumeStoreMode {
	return driver.LocalStore | driver.UseLocalMetaStore
}

// Create a local volume.
func (p *Local) Create(ctx driver.Context, id types.VolumeID) (*types.Volume, error) {
	ctx.Log.Debugf("Local create volume: %s", id.Name)

	var (
		mountPath = path.Join(dataDir, id.Name)
		size      string
	)

	// parse the mount path.
	if dir, ok := id.Options["mount"]; ok {
		mountPath = path.Join(dir, id.Name)
	}

	// parse the size.
	if value, ok := id.Options["opt.size"]; ok {
		sizeInt, err := bytefmt.ToMegabytes(value)
		if err != nil {
			return nil, err
		}
		size = strconv.Itoa(int(sizeInt)) + "M"
	}

	// create the volume path
	if st, exist := os.Stat(mountPath); exist != nil {
		if e := os.MkdirAll(mountPath, 0755); e != nil {
			return nil, e
		}
	} else if !st.IsDir() {
		return nil, fmt.Errorf("mount path is not a dir %s", mountPath)
	}

	return types.NewVolumeFromID(mountPath, size, id), nil
}

// Remove a local volume.
func (p *Local) Remove(ctx driver.Context, v *types.Volume) error {
	ctx.Log.Debugf("Local remove volume: %s", v.Name)
	mountPath := v.Path()

	if err := os.RemoveAll(mountPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %q directory failed, err: %v", mountPath, err)
	}

	return nil
}

// Path returns local volume's path.
func (p *Local) Path(ctx driver.Context, v *types.Volume) (string, error) {
	ctx.Log.Debugf("Local volume mount path: %s", v.Name)

	mountPath := v.Option("mount")
	if mountPath == "" {
		mountPath = dataDir
	}

	return path.Join(mountPath, v.Name), nil
}

// Options returns local volume's options.
func (p *Local) Options() map[string]types.Option {
	return map[string]types.Option{
		"mount": {Value: "", Desc: "local directory"},
	}
}

// Attach a local volume.
func (p *Local) Attach(ctx driver.Context, v *types.Volume) error {
	ctx.Log.Debugf("Local attach volume: %s", v.Name)
	mountPath := v.Path()
	size := v.Size()

	if st, exist := os.Stat(mountPath); exist != nil {
		if e := os.MkdirAll(mountPath, 0777); e != nil {
			return e
		}
	} else if !st.IsDir() {
		return fmt.Errorf("mount path is not a dir %s", mountPath)
	}

	if size != "" && size != "0" {
		quota.StartQuotaDriver(mountPath)
		if ex := quota.SetDiskQuota(mountPath, size, 0); ex != nil {
			return ex
		}
	}

	return nil
}

// Detach a local volume.
func (p *Local) Detach(ctx driver.Context, v *types.Volume) error {
	ctx.Log.Debugf("Local detach volume: %s", v.Name)

	return nil
}
