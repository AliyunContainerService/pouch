package tmpfs

import (
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"

	"strconv"
	"time"

	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/volume/driver"
	"github.com/alibaba/pouch/volume/types"
)

var (
	dataDir = "/etc/docker/plugins/tmpfs"
)

func init() {
	if err := driver.Register(&Tmpfs{}); err != nil {
		panic(err)
	}
}

// Tmpfs represents tmpfs volume driver.
type Tmpfs struct {
}

// Name returns tmpfs volume driver's name.
func (p *Tmpfs) Name(ctx driver.Context) string {
	return "tmpfs"
}

// StoreMode returns tmpfs volume driver's store mode.
func (p *Tmpfs) StoreMode(ctx driver.Context) driver.VolumeStoreMode {
	return driver.LocalStore | driver.UseLocalMetaStore
}

// Create a tmpfs volume.
func (p *Tmpfs) Create(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Tmpfs create volume: %s", v.Name)

	return nil
}

// Remove a tmpfs volume.
func (p *Tmpfs) Remove(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Tmpfs remove volume: %s", v.Name)

	return nil
}

// Path returns tmpfs volume's path.
func (p *Tmpfs) Path(ctx driver.Context, v *types.Volume) (string, error) {
	ctx.Log.Debugf("Tmpfs volume mount path: %s", v.Name)
	return path.Join(dataDir, v.Name), nil
}

// Options returns tmpfs volume's options.
func (p *Tmpfs) Options() map[string]types.Option {
	return map[string]types.Option{
		"ids":      {"", "tmpfs volume user's ids"},
		"reqID":    {"", "create tmpfs volume request id"},
		"freeTime": {"", "tmpfs volume free time"},
	}
}

// Attach a tmpfs volume.
func (p *Tmpfs) Attach(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Tmpfs attach volume: %s", v.Name)
	mountPath := v.Path()
	size := v.Size()
	reqID := v.Option("reqID")
	ids := v.Option("ids")

	if ids != "" {
		if !strings.Contains(ids, reqID) {
			ids = ids + "," + reqID
		}
	} else {
		ids = reqID
	}

	if err := os.MkdirAll(mountPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating %q directory: %v", mountPath, err)
	}

	if !utils.IsMountpoint(mountPath) {
		err := syscall.Mount("shm", mountPath, "tmpfs",
			uintptr(syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV),
			fmt.Sprintf("mode=1777,size=%s", size))
		if err != nil {
			return fmt.Errorf("mounting shm tmpfs: %s %v", mountPath, err)
		}
	}

	v.SetOption("ids", ids)
	v.SetOption("freeTime", "")

	return nil
}

// Detach a tmpfs volume.
func (p *Tmpfs) Detach(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Tmpfs detach volume: %s", v.Name)
	mountPath := v.Path()
	reqID := v.Option("reqID")
	ids := v.Option("ids")

	arr := strings.Split(ids, ",")
	newIDs := []string{}
	for _, id := range arr {
		if id != reqID {
			newIDs = append(newIDs, reqID)
		}
	}

	if len(newIDs) == 0 && utils.IsMountpoint(mountPath) {
		if err := syscall.Unmount(mountPath, 0); err != nil {
			return fmt.Errorf("failed to umount %q, err: %v", mountPath, err)
		}

		if err := os.Remove(mountPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %q directory failed, err: %v", mountPath, err)
		}

		v.SetOption("freeTime", strconv.FormatInt(time.Now().Unix(), 10))
	}

	v.SetOption("ids", strings.Join(newIDs, ","))

	return nil
}
