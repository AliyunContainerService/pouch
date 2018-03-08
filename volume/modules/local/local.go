// +build linux

package local

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/quota"
	"github.com/alibaba/pouch/volume/driver"
	"github.com/alibaba/pouch/volume/types"
)

var (
	dataDir = "/mnt/local"
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
func (p *Local) Create(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Local create volume: %s", v.Name)
	mountPath := v.Path()

	if st, exist := os.Stat(mountPath); exist != nil {
		if e := os.MkdirAll(mountPath, 0755); e != nil {
			return e
		}
	} else if !st.IsDir() {
		return fmt.Errorf("mount path is not a dir %s", mountPath)
	}

	return nil
}

// Remove a local volume.
func (p *Local) Remove(ctx driver.Context, v *types.Volume, s *types.Storage) error {
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
		"ids":      {Value: "", Desc: "local volume user's ids"},
		"reqID":    {Value: "", Desc: "create local volume request id"},
		"freeTime": {Value: "", Desc: "local volume free time"},
		"mount":    {Value: "", Desc: "local directory"},
	}
}

// Attach a local volume.
func (p *Local) Attach(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Local attach volume: %s", v.Name)
	mountPath := v.Path()
	size := v.Option("size")
	reqID := v.Option("reqID")
	ids := v.Option("ids")

	if ids != "" {
		if !strings.Contains(ids, reqID) {
			ids = ids + "," + reqID
		}
	} else {
		ids = reqID
	}

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

	v.SetOption("ids", ids)
	v.SetOption("freeTime", "")

	return nil
}

// Detach a local volume.
func (p *Local) Detach(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Local detach volume: %s", v.Name)
	reqID := v.Option("reqID")
	ids := v.Option("ids")

	arr := strings.Split(ids, ",")
	newIDs := []string{}
	for _, id := range arr {
		if id != reqID {
			newIDs = append(newIDs, reqID)
		}
	}

	if len(newIDs) == 0 {
		v.SetOption("freeTime", strconv.FormatInt(time.Now().Unix(), 10))
	}

	v.SetOption("ids", strings.Join(newIDs, ","))

	return nil
}
