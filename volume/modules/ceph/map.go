// +build linux

package ceph

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/volume/driver"
	"github.com/alibaba/pouch/volume/types"
)

var (
	bin          = "/usr/bin/rbd-nbd"
	poolName     = "rbd"
	objectSize   = "4M"
	imageFeature = "layering,exclusive-lock"
)

func getImageFeature(v *types.Volume) string {
	if s, ok := v.Spec.Extra["image-feature"]; ok {
		return s
	}
	return imageFeature
}

func getObjectSize(v *types.Volume) string {
	if s, ok := v.Spec.Extra["object-size"]; ok {
		return s
	}
	return objectSize
}

func getPool(v *types.Volume) string {
	if p, ok := v.Spec.Extra["pool"]; ok {
		return p
	}
	return poolName
}

func mkpool(poolName, volumeName string) string {
	return fmt.Sprintf("%s/%s", poolName, volumeName)
}

func rbdMap(ctx driver.Context, v *types.Volume, address, keyring string) (string, error) {
	poolName := getPool(v)
	name := v.Name

	args := []string{"map", mkpool(poolName, name), "--mon_host", address}
	exit, _, stderr, err := exec.RunWithRetry(3, 100*time.Microsecond, defaultTimeout, bin, args...)
	if err != nil || exit != 0 {
		err = fmt.Errorf("could not map %q: %d (%v) (%v)", name, exit, err, stderr)
		ctx.Log.Errorf("Ceph %s", err.Error())
		return "", err
	}

	var device string
	rbdmap, err := showMapped(ctx)
	if err != nil {
		return "", err
	}

	for _, rbd := range rbdmap.MapDevice {
		if rbd.Image == name && rbd.Pool == poolName {
			device = rbd.Device
			break
		}
	}
	if device == "" {
		return "", fmt.Errorf("volume %s in pool %s not found in RBD showmapped output", name, poolName)
	}

	ctx.Log.Infof("Ceph mapped volume %q as %q", name, device)

	return device, nil
}

func rbdUnmap(ctx driver.Context, v *types.Volume) error {
	err := exec.Retry(10, 100*time.Millisecond, func() error {
		rbdmap, err := showMapped(ctx)
		if err != nil {
			return err
		}

		err = doUnmap(ctx, v, rbdmap)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not unmap volume %q after 10 retries", v.Name)
	}

	return nil
}

func showMapped(ctx driver.Context) (RBDMap, error) {
	rbdmap := RBDMap{}

	exit, stdout, stderr, err := exec.RunWithRetry(3, 100*time.Millisecond, defaultTimeout,
		bin, "--dump-json", "list-mapped")
	if err != nil || exit != 0 {
		ctx.Log.Warnf("Ceph could not show mapped. Retrying: err:%v exit: %v, stderr: %v, stdout:%v",
			err, exit, stderr, stdout)
	} else if stdout != "" {
		err = json.Unmarshal([]byte(stdout), &rbdmap)
		if err != nil {
			ctx.Log.Errorf("Ceph could not parse rbd list-mapped output: %s", stdout)
		}
	}

	return rbdmap, err
}

func doUnmap(ctx driver.Context, v *types.Volume, rbdmap RBDMap) error {
	poolName := getPool(v)
	name := v.Name

	for _, rbd := range rbdmap.MapDevice {
		if rbd.Image != name || rbd.Pool != poolName {
			continue
		}
		ctx.Log.Debugf("Ceph unmapping volume %s/%s at device %q", poolName, name, strings.TrimSpace(rbd.Device))

		// Check device is exist or not.
		if _, err := os.Stat(rbd.Device); err != nil {
			ctx.Log.Errorf("Ceph trying to unmap device %q for %s/%s that does not exist, continuing",
				poolName, name, rbd.Device)
			continue
		}

		// Unmap device.
		exit, _, stderr, err := exec.Run(defaultTimeout, bin, "unmap", rbd.Device)
		if err != nil {
			ctx.Log.Errorf("Ceph could not unmap volume %q (device %q): %d (%v) (%s)",
				name, rbd.Device, exit, err, stderr)
			return err
		}
		if exit != 0 {
			err = fmt.Errorf("Ceph could not unmap volume %q (device %q): %d (%s)",
				name, rbd.Device, exit, stderr)
			ctx.Log.Error(err)
			return err
		}

		// Check have unmapped or not.
		rbdmap2, err := showMapped(ctx)
		if err != nil {
			return err
		}
		for _, rbd2 := range rbdmap2.MapDevice {
			if rbd.Image == rbd2.Image && rbd.Pool == rbd2.Pool {
				return fmt.Errorf("could not unmap volume %q, device: %q is exist",
					name, rbd.Image)
			}
		}
		break
	}
	return nil
}
