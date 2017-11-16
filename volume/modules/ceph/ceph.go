// +build linux

package ceph

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/volume/driver"
	"github.com/alibaba/pouch/volume/types"
	"github.com/alibaba/pouch/volume/types/meta"

	"github.com/go-ini/ini"
)

func init() {
	if err := driver.Register(&Ceph{}); err != nil {
		panic(err)
	}
}

// Ceph represents ceph volume driver struct.
type Ceph struct{}

// Name returns volume driver's name.
func (c *Ceph) Name(ctx driver.Context) string {
	return "ceph"
}

// StoreMode returns volume driver's store mode.
func (c *Ceph) StoreMode(ctx driver.Context) driver.VolumeStoreMode {
	return driver.RemoteStore | driver.CreateDeleteInCentral
}

// Create a ceph volume.
func (c *Ceph) Create(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Ceph create volume: %s", v.Name)

	poolName := getPool(v)
	objSize := getObjectSize(v)
	feature := getImageFeature(v)
	name := v.Name
	conf, ok := ctx.GetString("conf")
	if !ok {
		return fmt.Errorf("get config file fail")
	}

	var cmdList []string
	base := v.Option("base")
	if base != "" {
		cmdList = []string{"clone", fmt.Sprintf("%s/%s@snap", poolName, base),
			fmt.Sprintf("%s/%s", poolName, name)}
	} else {
		cmdList = []string{"create", name, "--pool", poolName, "--size", v.Size(),
			"--object-size", objSize, "--image-feature", feature}
	}

	ctx.Log.Infof("Ceph create volume, command: %v", cmdList)

	cmd := NewCephCommand("rbd", conf)
	err := cmd.RunCommand(nil, cmdList...)
	if err != nil {
		ctx.Log.Errorf("Ceph rbd create failed, err: %v", err)
		return err
	}

	return nil
}

// Remove a ceph volume.
func (c *Ceph) Remove(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Ceph Remove volume: %s", v.Name)

	poolName := getPool(v)
	name := v.Name
	conf, ok := ctx.GetString("conf")
	if !ok {
		return fmt.Errorf("get config file fail")
	}

	cmdList := []string{"snap", "purge", mkpool(poolName, name)}

	ctx.Log.Infof("Ceph remove volume, command: %v", cmdList)

	cmd := NewCephCommand("rbd", conf)
	err := cmd.RunCommand(nil, cmdList...)
	if err != nil {
		return err
	}

	cmdList = []string{"rm", mkpool(poolName, name)}
	err = cmd.RunCommand(nil, cmdList...)
	if err != nil {
		return err
	}

	return nil
}

// Path returns ceph volume mount path.
func (c *Ceph) Path(ctx driver.Context, v *types.Volume) (string, error) {
	ctx.Log.Debugf("Ceph volume mount path: %s", v.Name)
	return path.Join("/mnt", getPool(v), v.Name), nil
}

// Options returns ceph volume options.
func (c *Ceph) Options() map[string]types.Option {
	return map[string]types.Option{
		"conf":          {Value: "/etc/ceph/ceph.conf", Desc: "set ceph config file"},
		"key":           {Value: "", Desc: "set ceph keyring file"},
		"pool":          {Value: "rbd", Desc: "set rbd image pool"},
		"image-feature": {Value: "", Desc: "set rbd image feature"},
		"object-size":   {Value: "4M", Desc: "set rbd image object size"},
		"base":          {Value: "", Desc: "create image by base image"},
	}
}

// Format is used to format ceph volume.
func (c *Ceph) Format(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Ceph format volume: %s", v.Name)

	device, err := rbdMap(ctx, v, s.Spec.Address, s.Spec.Key)
	if err != nil {
		ctx.Log.Errorf("Ceph map volume: %s failed: %v", v.Name, err)
		return err
	}

	fs := v.FileSystem()
	ctx.Log.Infof("Ceph make file system, volume: %s, device: %s, fs: %v", v.Name, device, fs)

	if err := utils.MakeFSVolume(fs, device, defaultFormatTimeout); err != nil {
		ctx.Log.Errorf("Ceph mkfs error: %v", err)
		if err := rbdUnmap(ctx, v); err != nil {
			ctx.Log.Errorf("Ceph error while trying to unmap after failed filesystem creation: %v", err)
		}
		return err
	}

	ctx.Log.Infof("Ceph start to unmap %s", v.Name)
	if err := rbdUnmap(ctx, v); err != nil {
		ctx.Log.Errorf("Ceph unmap v :%s err: %v", v.Name, err)
		return err
	}

	return nil
}

// Attach is used to attach ceph volume.
func (c *Ceph) Attach(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Ceph attach volume: %s", v.Name)
	var err error
	volumePath := v.Path()

	// Map rbd device
	devName, err := rbdMap(ctx, v, s.Spec.Address, s.Spec.Key)
	if err != nil {
		ctx.Log.Errorf("Ceph volume: %s map error: %v", v.Name, err)
		return err
	}

	// Create directory to mount
	err = os.MkdirAll(volumePath, 0755)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating %q directory: %v", volumePath, err)
	}

	// Mount the RBD
	mountOpt := v.MountOption()
	ctx.Log.Infof("Ceph mount volume: %s, device: %s, mountpath: %s, opt: %v",
		v.Name, devName, volumePath, mountOpt)

	err = utils.MountVolume(mountOpt, devName, volumePath, time.Second*60)
	if err != nil {
		ctx.Log.Errorf("Ceph volume: %s mount failed, err: %v", v.Name, err)
		return fmt.Errorf("failed to mount rbd device %q: %v", devName, err)
	}

	return nil
}

// Detach is used to detach ceph volume.
func (c *Ceph) Detach(ctx driver.Context, v *types.Volume, s *types.Storage) error {
	ctx.Log.Debugf("Ceph detach volume: %s", v.Name)
	mountPath := v.Path()

	err := exec.Retry(3, 3*time.Second, func() error {
		if err := syscall.Unmount(mountPath, 0); err != nil {
			ctx.Log.Errorf("Ceph volume: %s unmount %q failed, err: %v", v.Name, mountPath, err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Remove the mounted directory
	// FIXME remove all, but only after the FIXME above.
	err = exec.Retry(3, 3*time.Second, func() error {
		if err := os.Remove(mountPath); err != nil && !os.IsNotExist(err) {
			ctx.Log.Errorf("Ceph removing %q directory failed, err: %v", mountPath, err)
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to remove mount path: %s after 3 retries", mountPath)
	}

	// Unmap rbd device
	if err := rbdUnmap(ctx, v); err != os.ErrNotExist {
		return err
	}

	return nil
}

// Report is used to collect ceph storage information and report to control server.
func (c *Ceph) Report(ctx driver.Context) ([]*types.Storage, error) {
	var err error
	var storages []*types.Storage
	conf, ok := ctx.GetString("conf")
	if !ok {
		return nil, fmt.Errorf("get config file fail")
	}
	keyFile, ok := ctx.GetString("keyring")
	if !ok {
		return nil, fmt.Errorf("get keyring file fail")
	}

	s := &types.Storage{
		ObjectMeta: meta.ObjectMeta{
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec:   &types.StorageSpec{},
		Status: &types.StorageStatus{},
	}

	// analyze keyfile
	cfg, err := ini.Load(keyFile)
	if err != nil {
		ctx.Log.Errorf("Ceph load keyring error : %v", err)
		return nil, err
	}
	key, err := cfg.Section("client.admin").GetKey("key")
	if err != nil {
		ctx.Log.Errorf("Ceph get client.admin section error: %v", err)
		return nil, err
	}
	keyring := fmt.Sprintf("%s:%s", "admin", key.Value())
	encoded := base64.StdEncoding.EncodeToString([]byte(keyring))
	s.Spec.Key = encoded

	// get ceph quorum status
	cmd := NewCephCommand("ceph", conf)
	status := &QuorumStatus{}
	if err = cmd.RunCommand(status, QuorumCommand...); err != nil {
		ctx.Log.Errorf("Ceph run quorum error: %v", err)
		return nil, err
	}
	monAddrs := []string{}
	for _, mon := range status.MonMap.Mons {
		monAddrs = append(monAddrs, strings.Split(mon.Addr, "/")[0])
	}
	s.Spec.Address = strings.Join(monAddrs, ",")

	ctx.Log.Debugf("Ceph report storage: %s", s.Spec.Address)

	// get ceph status
	stats := &Stats{}
	if err = cmd.RunCommand(stats, CephStatusCommand...); err != nil {
		ctx.Log.Errorf("Ceph run cephstatus command error: %v", err)
		return nil, err
	}
	s.Status.Schedulable = stats.Health.OverallStatus != HealthErr
	s.Status.HealthyStatus = stats.Health.OverallStatus

	// get ceph usage
	usage := &PoolStats{}
	if err = cmd.RunCommand(usage, CephUsageCommand...); err != nil {
		return nil, err
	}
	poolSpec := make(map[string]types.PoolSpec)
	for _, p := range usage.Pools {
		poolSpec[p.Name] = types.PoolSpec{
			Capacity:  p.Stats.BytesUsed + p.Stats.MaxAvail,
			Available: p.Stats.MaxAvail,
		}
	}
	s.Spec.PoolSpec = poolSpec

	s.UID = stats.ID

	storages = append(storages, s)

	return storages, nil
}
