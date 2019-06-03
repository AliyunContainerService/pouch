package mgr

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/alibaba/pouch/apis/opts"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/archive"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/randomid"
	"github.com/alibaba/pouch/pkg/system"
	"github.com/alibaba/pouch/storage/quota"
	volumetypes "github.com/alibaba/pouch/storage/volume/types"

	"github.com/containerd/containerd/mount"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (mgr *ContainerManager) attachVolume(ctx context.Context, name string, c *Container) (string, string, error) {
	driver := volumetypes.DefaultBackend
	v, err := mgr.VolumeMgr.Get(ctx, name)
	if err != nil || v == nil {
		opts := map[string]string{
			"backend": driver,
		}
		if _, err := mgr.VolumeMgr.Create(ctx, name, c.HostConfig.VolumeDriver, opts, nil); err != nil {
			logrus.Errorf("failed to create volume(%s), err(%v)", name, err)
			return "", "", errors.Wrap(err, "failed to create volume")
		}
	} else {
		driver = v.Driver()
	}

	if _, err := mgr.VolumeMgr.Attach(ctx, name, map[string]string{volumetypes.OptionRef: c.ID}); err != nil {
		logrus.Errorf("failed to attach volume(%s), err(%v)", name, err)
		return "", "", errors.Wrap(err, "failed to attach volume")
	}

	mountPath, err := mgr.VolumeMgr.Path(ctx, name)
	if err != nil {
		logrus.Errorf("failed to get the mount path of volume(%s), err(%v)", name, err)
		return "", "", errors.Wrap(err, "failed to get volume mount path")
	}

	return mountPath, driver, nil
}

func (mgr *ContainerManager) generateMountPoints(ctx context.Context, c *Container) error {
	var err error

	if c.Config.Volumes == nil {
		c.Config.Volumes = make(map[string]interface{})
	}

	if c.Mounts == nil {
		c.Mounts = make([]*types.MountPoint, 0)
	}

	// define a volume map to duplicate removal
	volumeSet := map[string]struct{}{}

	defer func() {
		if err != nil {
			if err := mgr.detachVolumes(ctx, c, false); err != nil {
				logrus.Errorf("failed to detach volume, err(%v)", err)
			}
		}
	}()

	// 1. read MountPoints from other containers
	err = mgr.getMountPointFromContainers(ctx, c, volumeSet)
	if err != nil {
		return errors.Wrap(err, "failed to get mount point from containers")
	}

	// 2. read MountPoints from binds
	err = mgr.getMountPointFromBinds(ctx, c, volumeSet)
	if err != nil {
		return errors.Wrap(err, "failed to get mount point from binds")
	}

	// 3. read MountPoints from image
	err = mgr.getMountPointFromImage(ctx, c, volumeSet)
	if err != nil {
		return errors.Wrap(err, "failed to get mount point from image")
	}

	// 4. read MountPoints from Config.Volumes
	err = mgr.getMountPointFromVolumes(ctx, c, volumeSet)
	if err != nil {
		return errors.Wrap(err, "failed to get mount point from volumes")
	}

	// do filepath clean
	for i := range c.Mounts {
		m := c.Mounts[i]
		m.Source = filepath.Clean(m.Source)
		m.Destination = filepath.Clean(m.Destination)
	}

	// populate the volumes
	err = mgr.populateVolumes(ctx, c)
	if err != nil {
		return errors.Wrap(err, "failed to populate volumes")
	}

	return nil
}

func (mgr *ContainerManager) getMountPointFromBinds(ctx context.Context, c *Container, volumeSet map[string]struct{}) error {
	var err error

	logrus.Debugf("bind volumes(%v)", c.HostConfig.Binds)

	// parse binds
	for _, b := range c.HostConfig.Binds {
		var parts []string
		parts, err = opts.CheckBind(b)
		if err != nil {
			return err
		}

		mode := ""
		mp := new(types.MountPoint)

		switch len(parts) {
		case 1:
			mp.Source = ""
			mp.Destination = parts[0]
		case 2:
			mp.Source = parts[0]
			mp.Destination = parts[1]
			mp.Named = true
		case 3:
			mp.Source = parts[0]
			mp.Destination = parts[1]
			mode = parts[2]
			mp.Named = true
		default:
			return errors.Errorf("unknown bind(%s)", b)
		}

		if opts.CheckDuplicateMountPoint(c.Mounts, mp.Destination) {
			logrus.Warnf("duplicate mountpoint(%s)", mp.Destination)
			continue
		}

		if mp.Source == "" {
			mp.Source = randomid.Generate()
		}

		err = opts.ParseBindMode(mp, mode)
		if err != nil {
			logrus.Errorf("failed to parse bind mode(%s), err(%v)", mode, err)
			return err
		}

		if !path.IsAbs(mp.Source) {
			// volume bind.
			name := mp.Source
			if _, exist := volumeSet[name]; !exist {
				_, mp.Driver, err = mgr.attachVolume(ctx, name, c)
				if err != nil {
					logrus.Errorf("failed to bind volume(%s), err(%v)", name, err)
					return errors.Wrap(err, "failed to bind volume")
				}

				volumeSet[name] = struct{}{}
			}

			volume, err := mgr.VolumeMgr.Get(ctx, name)
			if err != nil || volume == nil {
				logrus.Errorf("failed to get volume(%s), err(%v)", name, err)
				return errors.Wrapf(err, "failed to get volume(%s)", name)
			}
			mp.Driver = volume.Driver()
			mp.Name = name

			mp.Source, err = mgr.VolumeMgr.Path(ctx, name)
			if err != nil {
				return err
			}

			if mp.Replace != "" {
				switch mp.Replace {
				case "dr":
					mp.Source = path.Join(mp.Source, mp.Destination)
				case "rr":
					mp.Source = path.Join(mp.Source, randomid.Generate())
				}

				mp.Name = ""
				mp.Named = false
				mp.Driver = ""
			}
		} else {
			mp.CopyData = false
		}

		c.Mounts = append(c.Mounts, mp)
	}

	return nil
}

func (mgr *ContainerManager) getMountPointFromVolumes(ctx context.Context, c *Container, volumeSet map[string]struct{}) error {
	var err error

	// parse volumes
	for dest := range c.Config.Volumes {
		if opts.CheckDuplicateMountPoint(c.Mounts, dest) {
			logrus.Warnf("duplicate mountpoint(%s) from volumes", dest)
			continue
		}

		// check if volume has been created
		name := randomid.Generate()
		if _, exist := volumeSet[name]; exist {
			continue
		}

		mp := new(types.MountPoint)
		mp.Name = name
		mp.Destination = dest

		mp.Source, mp.Driver, err = mgr.attachVolume(ctx, mp.Name, c)
		if err != nil {
			logrus.Errorf("failed to bind volume(%s), err(%v)", mp.Name, err)
			return errors.Wrap(err, "failed to bind volume")
		}

		err = opts.ParseBindMode(mp, "")
		if err != nil {
			logrus.Errorf("failed to parse mode, err(%v)", err)
			return err
		}

		volumeSet[mp.Name] = struct{}{}
		c.Mounts = append(c.Mounts, mp)
	}

	return nil
}

func (mgr *ContainerManager) getMountPointFromImage(ctx context.Context, c *Container, volumeSet map[string]struct{}) error {
	var err error

	// parse volumes from image
	image, err := mgr.ImageMgr.GetImage(ctx, c.Image)
	if err != nil {
		return errors.Wrapf(err, "failed to get image(%s)", c.Image)
	}
	for dest := range image.Config.Volumes {
		// check if volume has been created
		name := randomid.Generate()
		if _, exist := volumeSet[name]; exist {
			continue
		}

		if opts.CheckDuplicateMountPoint(c.Mounts, dest) {
			logrus.Warnf("duplicate mountpoint(%s) from image", dest)
			continue
		}

		mp := new(types.MountPoint)
		mp.Name = name
		mp.Destination = dest

		mp.Source, mp.Driver, err = mgr.attachVolume(ctx, mp.Name, c)
		if err != nil {
			logrus.Errorf("failed to bind volume(%s), err(%v)", mp.Name, err)
			return errors.Wrap(err, "failed to bind volume")
		}

		err = opts.ParseBindMode(mp, "")
		if err != nil {
			logrus.Errorf("failed to parse mode, err(%v)", err)
			return err
		}

		volumeSet[mp.Name] = struct{}{}
		c.Mounts = append(c.Mounts, mp)
	}

	return nil
}

func (mgr *ContainerManager) getMountPointFromContainers(ctx context.Context, container *Container, volumeSet map[string]struct{}) error {
	var err error

	// parse volumes from other containers
	for _, v := range container.HostConfig.VolumesFrom {
		var containerID, mode string
		containerID, mode, err = opts.ParseVolumesFrom(v)
		if err != nil {
			return err
		}

		var oldContainer *Container
		oldContainer, err = mgr.Get(ctx, containerID)
		if err != nil {
			return err
		}

		for _, oldMountPoint := range oldContainer.Mounts {
			if opts.CheckDuplicateMountPoint(container.Mounts, oldMountPoint.Destination) {
				logrus.Warnf("duplicate mount point %s on container %s", oldMountPoint.Destination, containerID)
				continue
			}

			mp := &types.MountPoint{
				Source:      oldMountPoint.Source,
				Destination: oldMountPoint.Destination,
				Driver:      oldMountPoint.Driver,
				Named:       oldMountPoint.Named,
				Mode:        oldMountPoint.Mode,
				Replace:     oldMountPoint.Replace,
				Propagation: oldMountPoint.Propagation,
			}

			if _, exist := volumeSet[oldMountPoint.Name]; len(oldMountPoint.Name) > 0 && !exist {
				mp.Name = oldMountPoint.Name
				mp.Source, mp.Driver, err = mgr.attachVolume(ctx, oldMountPoint.Name, container)
				if err != nil {
					logrus.Errorf("failed to bind volume(%s), err(%v)", oldMountPoint.Name, err)
					return errors.Wrap(err, "failed to bind volume")
				}

				volumeSet[mp.Name] = struct{}{}
			}

			err = opts.ParseBindMode(mp, mode)
			if err != nil {
				logrus.Errorf("failed to parse volumes-from mode(%s), err(%v)", mode, err)
				return err
			}

			// if mode has no set rw mode, so need to inherit old mountpoint rw mode.
			if !(strings.Contains(mode, "ro") || strings.Contains(mode, "rw")) {
				mp.RW = oldMountPoint.RW
			}

			// if mode has no set copy mode, so need to inherit old mountpoint copy mode.
			if !strings.Contains(mode, "nocopy") {
				if !strings.Contains(oldMountPoint.Mode, "nocopy") {
					mp.CopyData = true
				} else {
					mp.CopyData = false
				}
			}

			container.Mounts = append(container.Mounts, mp)
		}
	}

	return nil
}

func (mgr *ContainerManager) populateVolumes(ctx context.Context, c *Container) error {
	// sort mounts by destination directory string shortest length.
	// the reason is: there are two mounts: /home/admin and /home/admin/log,
	// when do copy data with dr mode, if the data of /home/admin/log is copied first,
	// it will cause /home/admin don't copy data since the destination directory is not empty.
	c.Mounts = sortMountPoint(c.Mounts)

	for _, mp := range c.Mounts {
		if mp.Driver == "tmpfs" {
			continue
		}

		if !mp.CopyData {
			continue
		}

		logrus.Debugf("copying image data from (%s:%s), to volume(%s) or path(%s)",
			c.ID, mp.Destination, mp.Name, mp.Source)

		imagePath := path.Join(c.MountFS, mp.Destination)

		err := copyImageContent(imagePath, mp.Source)
		if err != nil {
			logrus.Errorf("failed to copy image contents, volume[imagepath(%s), source(%s)], err(%v)", imagePath, mp.Source, err)
			return errors.Wrapf(err, "failed to copy image content, image(%s), host(%s)", imagePath, mp.Source)
		}
	}

	for _, mp := range c.Mounts {
		if _, err := os.Stat(mp.Source); err != nil {
			// host directory bind into container.
			if !os.IsNotExist(err) {
				return errors.Wrapf(err, "failed to stat %q", mp.Source)
			}
			// Create the host path if it doesn't exist.
			if err = os.MkdirAll(mp.Source, 0755); err != nil {
				return errors.Wrapf(err, "failed to mkdir %q", mp.Source)
			}
		}
	}

	return nil
}

func (mgr *ContainerManager) setMountTab(ctx context.Context, c *Container) error {
	logrus.Debugf("start to set mount tab into container")

	if len(c.MountFS) == 0 {
		return nil
	}

	// set rootfs mount tab
	context := "/ / ext4 rw 0 0\n"
	if rootID, e := system.GetDevID(c.MountFS); e == nil {
		_, _, rootFsType := quota.CheckMountpoint(rootID)
		if len(rootFsType) > 0 {
			context = fmt.Sprintf("/ / %s rw 0 0\n", rootFsType)
		}
	}

	// set mount point tab
	i := 1
	for _, m := range c.Mounts {
		if m.Source == "" || m.Destination == "" {
			continue
		}

		finfo, err := os.Stat(m.Source)
		if err != nil || !finfo.IsDir() {
			continue
		}

		tempLine := fmt.Sprintf("/dev/v%02dd %s ext4 rw 0 0\n", i, m.Destination)
		if tmpID, e := system.GetDevID(m.Source); e == nil {
			_, _, fsType := quota.CheckMountpoint(tmpID)
			if len(fsType) > 0 {
				tempLine = fmt.Sprintf("/dev/v%02dd %s %s rw 0 0\n", i, m.Destination, fsType)
			}
		}

		i++
		context += tempLine
	}

	// set shm mount tab
	context += "shm /dev/shm tmpfs rw 0 0\n"

	// save into mtab file.
	mtabPath := filepath.Join(c.MountFS, "etc/mtab")

	// make directory, $BaseFS/etc/
	os.MkdirAll(path.Dir(mtabPath), 0755)
	// remove it before modify
	os.Remove(mtabPath)

	err := ioutil.WriteFile(mtabPath, []byte(context), 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write file(%s)", mtabPath)
	}

	logrus.Infof("write mtab file to (%s)", mtabPath)

	return nil
}

func (mgr *ContainerManager) getDiskQuotaMountPoints(ctx context.Context, c *Container, mounted bool) ([]*types.MountPoint, error) {
	var mounts []*types.MountPoint

	for _, mp := range c.Mounts {
		// skip volume mount or replace mode mount
		if mp.Replace != "" || mp.Source == "" || mp.Destination == "" {
			logrus.Debugf("skip volume mount or replace mode mount")
			continue
		}

		// skip volume that has set size
		if mp.Name != "" {
			v, err := mgr.VolumeMgr.Get(ctx, mp.Name)
			if err != nil {
				logrus.Warnf("failed to get volume(%s)", mp.Name)
				continue
			}

			if v.Size() != "" {
				logrus.Debugf("skip volume(%s) with size", mp.Name)
				continue
			}
		}

		// skip non-directory path.
		if fd, err := os.Stat(mp.Source); err != nil || !fd.IsDir() {
			logrus.Debugf("skip non-directory path(%s)", mp.Source)
			continue
		}

		mounts = append(mounts, mp)
	}

	// add rootfs mountpoint
	rootfs, err := mgr.getRootfs(ctx, c, mounted)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get rootfs")
	}
	mounts = append(mounts, &types.MountPoint{
		Source:      rootfs,
		Destination: "/",
	})

	return mounts, nil
}

func checkDupQuotaMap(qms []*quota.QMap, qm *quota.QMap) *quota.QMap {
	for _, prev := range qms {
		if qm.Expression != "" && qm.Expression == prev.Expression {
			return prev
		}
	}
	return nil
}

func (mgr *ContainerManager) setDiskQuota(ctx context.Context, c *Container, mounted bool) error {
	var (
		err           error
		globalQuotaID uint32
	)

	// get default quota
	quotas := c.Config.DiskQuota

	if quota.IsSetQuotaID(c.Config.QuotaID) {
		id, err := strconv.Atoi(c.Config.QuotaID)
		if err != nil {
			return errors.Wrapf(err, "invalid argument, QuotaID(%s)", c.Config.QuotaID)
		}

		// if QuotaID is < 0, it means pouchd alloc a unique quota id.
		if id < 0 {
			globalQuotaID, err = quota.GetNextQuotaID()
			if err != nil {
				return errors.Wrap(err, "failed to get next quota id")
			}

			// update QuotaID
			c.Config.QuotaID = strconv.Itoa(int(globalQuotaID))
		} else {
			globalQuotaID = uint32(id)
		}
	}

	// get mount points that can set disk quota.
	mounts, err := mgr.getDiskQuotaMountPoints(ctx, c, mounted)
	if err != nil {
		return errors.Wrap(err, "failed to get mount point that can set disk quota")
	}

	var qms []*quota.QMap
	for _, mp := range mounts {
		// get quota size
		var (
			found bool
			qm    *quota.QMap
		)
		for exp, size := range quotas {
			if strings.Contains(exp, "&") {
				for _, p := range strings.Split(exp, "&") {
					if p == mp.Destination {
						found = true
						qm = &quota.QMap{
							Source:      mp.Source,
							Destination: mp.Destination,
							Expression:  exp,
							Size:        size,
						}
						break
					}
				}
			} else {
				re := regexp.MustCompile(exp)

				findStr := re.FindString(mp.Destination)
				if findStr == mp.Destination {
					qm = &quota.QMap{
						Source:      mp.Source,
						Destination: mp.Destination,
						Size:        size,
					}
					if exp != ".*" {
						break
					}
				}
			}

			if found {
				break
			}
		}

		if qm != nil {
			// check duplicate quota map
			prev := checkDupQuotaMap(qms, qm)
			if prev == nil {
				// get new quota id
				id := globalQuotaID
				if id == 0 {
					id, err = quota.GetNextQuotaID()
					if err != nil {
						return errors.Wrap(err, "failed to get next quota id")
					}
				}
				qm.QuotaID = id
			} else {
				qm.QuotaID = prev.QuotaID
			}

			qms = append(qms, qm)
		}
	}

	// make quota effective
	for _, qm := range qms {
		if qm.Destination == "/" {
			// set rootfs quota
			_, err = quota.SetRootfsDiskQuota(qm.Source, qm.Size, qm.QuotaID)
			if err != nil {
				logrus.Warnf("failed to set rootfs quota, mountfs(%s), size(%s), quota id(%d), err(%v)",
					qm.Source, qm.Size, qm.QuotaID, err)
			}
		} else {
			err := quota.SetDiskQuota(qm.Source, qm.Size, qm.QuotaID)
			if err != nil {
				logrus.Warnf("failed to set disk quota, directory(%s), size(%s), quota id(%d), err(%v)",
					qm.Source, qm.Size, qm.QuotaID, err)
			}
		}
	}

	return nil
}

func (mgr *ContainerManager) detachVolumes(ctx context.Context, c *Container, remove bool) error {
	for _, mount := range c.Mounts {
		name := mount.Name
		if name == "" {
			continue
		}

		_, err := mgr.VolumeMgr.Detach(ctx, name, map[string]string{volumetypes.OptionRef: c.ID})
		if err != nil {
			logrus.Warnf("failed to detach volume(%s), err(%v)", name, err)
		}

		if remove && !mount.Named {
			if err := mgr.VolumeMgr.Remove(ctx, name); err != nil && !errtypes.IsInUse(err) {
				logrus.Warnf("failed to remove volume(%s) when remove container", name)
			}
		}
	}

	return nil
}

func (mgr *ContainerManager) attachVolumes(ctx context.Context, c *Container) (err0 error) {
	rollbackVolumes := make([]string, 0, len(c.Mounts))

	defer func() {
		if err0 != nil {
			for _, name := range rollbackVolumes {
				if _, err := mgr.VolumeMgr.Detach(ctx, name, map[string]string{volumetypes.OptionRef: c.ID}); err != nil {
					logrus.Warnf("[rollback] failed to detach volume(%s), err(%v)", name, err)
				}
			}
		}
	}()

	for _, mount := range c.Mounts {
		name := mount.Name
		if name == "" {
			continue
		}

		_, err := mgr.VolumeMgr.Attach(ctx, name, map[string]string{volumetypes.OptionRef: c.ID})
		if err != nil {
			logrus.Warnf("failed to attach volume(%s), err(%v)", name, err)
			return err
		}

		rollbackVolumes = append(rollbackVolumes, name)
	}
	return nil
}

// setMountFS is used to set mountfs directory.
func (mgr *ContainerManager) setMountFS(ctx context.Context, c *Container) {
	c.MountFS = path.Join(mgr.Store.Path(c.ID), "rootfs")
}

// Mount sets the container rootfs
// preCreate = false, mount rootfs at c.BaseFS
// preCreate = true, pouchd does some initial job before container created, so we mount rootfs at c.MountFS
func (mgr *ContainerManager) Mount(ctx context.Context, c *Container, preCreate bool) error {

	mounts, err := mgr.Client.GetMounts(ctx, c.ID)
	if err != nil {
		return err
	} else if len(mounts) != 1 {
		return fmt.Errorf("failed to get snapshot %s mounts: not equals 1", c.ID)
	}

	rootfs := c.BaseFS

	if preCreate {
		if c.MountFS == "" {
			mgr.setMountFS(ctx, c)
		}
		rootfs = c.MountFS
	}

	err = os.MkdirAll(rootfs, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return mounts[0].Mount(rootfs)
}

// Unmount unsets the container rootfs
// cleanup decides whether to clean up the dir or not
func (mgr *ContainerManager) Unmount(ctx context.Context, c *Container, preCreate bool, cleanup bool) error {
	// TODO: if umount is failed, and how to deal it.
	rootfs := c.MountFS
	if !preCreate {
		rootfs = c.BaseFS
	}
	var err error
	err = mount.Unmount(rootfs, 0)

	if err != nil {
		return errors.Wrapf(err, "failed to umount rootfs(%s)", rootfs)
	}

	if cleanup {
		if preCreate {
			return os.RemoveAll(rootfs)
		}
		// also need to remove dir named by container ID
		cleanPath, _ := filepath.Split(rootfs)
		logrus.Debugf("clean path of unmount is %s", cleanPath)
		return os.RemoveAll(cleanPath)
	}

	return err
}

func (mgr *ContainerManager) initContainerStorage(ctx context.Context, c *Container) (err error) {
	if err = mgr.Mount(ctx, c, true); err != nil {
		return errors.Wrapf(err, "failed to mount rootfs(%s)", c.MountFS)
	}

	defer func() {
		if umountErr := mgr.Unmount(ctx, c, true, true); umountErr != nil {
			if err != nil {
				err = errors.Wrapf(err, "failed to umount rootfs(%s), err(%v)", c.MountFS, umountErr)
			} else {
				err = errors.Wrapf(umountErr, "failed to umount rootfs(%s)", c.MountFS)
			}
		}
	}()

	// parse volume config
	if err = mgr.generateMountPoints(ctx, c); err != nil {
		return errors.Wrap(err, "failed to parse volume argument")
	}

	// try to setup container working directory
	if err := mgr.SetupWorkingDirectory(ctx, c); err != nil {
		return errors.Wrapf(err, "failed to setup container %s working directory", c.ID)
	}

	// set mount point disk quota
	if err = mgr.setDiskQuota(ctx, c, true); err != nil {
		// just ignore failed to set disk quota
		logrus.Warnf("failed to set disk quota, err(%v)", err)
	}

	// set volumes into /etc/mtab in container
	if err = mgr.setMountTab(ctx, c); err != nil {
		return errors.Wrap(err, "failed to set mount tab")
	}

	return nil
}

// SetupWorkingDirectory setup working directory for container
func (mgr *ContainerManager) SetupWorkingDirectory(ctx context.Context, c *Container) error {
	if c.Config.WorkingDir == "" {
		return nil
	}

	if c.MountFS == "" {
		mgr.setMountFS(ctx, c)
	}

	workingDir := filepath.Clean(c.Config.WorkingDir)
	resourcePath := c.GetResourcePath(c.MountFS, workingDir)

	// TODO(ziren): not care about File mode
	err := os.MkdirAll(resourcePath, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

func (mgr *ContainerManager) getRootfs(ctx context.Context, c *Container, mounted bool) (string, error) {
	var (
		rootfs string
		err    error
	)
	if c.IsRunningOrPaused() && c.Snapshotter != nil {
		basefs, ok := c.Snapshotter.Data["MergedDir"]
		if !ok || basefs == "" {
			return "", fmt.Errorf("container is running, but MergedDir is missing")
		}
		rootfs = basefs
	} else if !mounted {
		if err = mgr.Mount(ctx, c, true); err != nil {
			return "", errors.Wrapf(err, "failed to mount rootfs: (%s)", c.MountFS)
		}
		rootfs = c.MountFS

		defer func() {
			if err = mgr.Unmount(ctx, c, true, true); err != nil {
				logrus.Errorf("failed to umount rootfs: (%s), err: (%v)", c.MountFS, err)
			}
		}()
	} else {
		rootfs = c.MountFS
	}

	return rootfs, nil
}

func sortMountPoint(mounts []*types.MountPoint) []*types.MountPoint {
	sort.Slice(mounts, func(i, j int) bool {
		if len(mounts[i].Destination) < len(mounts[j].Destination) {
			return true
		}
		return false
	})

	return mounts
}

func copyImageContent(source, destination string) error {
	fi, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	} else if !fi.IsDir() {
		return nil
	}

	fi, err = os.Stat(destination)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// destination directory is not exist, so mkdir for it.
		logrus.Warnf("(%s) is not exist", destination)
		if err := os.MkdirAll(destination, 0755); err != nil && !os.IsExist(err) {
			return err
		}
	} else if !fi.IsDir() {
		return nil
	}

	return copyExistContents(source, destination)
}

func copyExistContents(source, destination string) error {
	volList, err := ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	if len(volList) > 0 {
		dstList, err := ioutil.ReadDir(destination)
		if err != nil {
			return err
		}
		if len(dstList) == 0 {
			logrus.Debugf("copy (%s) to (%s) with tar", source, destination)
			err := archive.CopyWithTar(source, destination)
			if err != nil {
				logrus.Errorf("copyImageContent: %v", err)
				return err
			}
		}
	}
	return copyOwnership(source, destination)
}

func copyOwnership(source, destination string) error {
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed to get file %s system info", source)
	}

	err = os.Chown(destination, int(sys.Uid), int(sys.Gid))
	if err != nil {
		return err
	}

	return os.Chmod(destination, os.FileMode(fi.Mode()))
}
