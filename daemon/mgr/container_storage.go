package mgr

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

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
			logrus.Errorf("failed to create volume(%s): %v", name, err)
			return "", "", errors.Wrap(err, "failed to create volume")
		}
	} else {
		driver = v.Driver()
	}

	if _, err := mgr.VolumeMgr.Attach(ctx, name, map[string]string{volumetypes.OptionRef: c.ID}); err != nil {
		logrus.Errorf("failed to attach volume(%s):: %v", name, err)
		return "", "", errors.Wrap(err, "failed to attach volume")
	}

	mountPath, err := mgr.VolumeMgr.Path(ctx, name)
	if err != nil {
		logrus.Errorf("failed to get the mount path of volume(%s): %v", name, err)
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
				logrus.Errorf("failed to detach volume: %v", err)
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

	// set volumes into /etc/mtab in container
	err = mgr.setMountTab(ctx, c)
	if err != nil {
		return errors.Wrap(err, "failed to set mount tab")
	}

	return nil
}

func (mgr *ContainerManager) getMountPointFromBinds(ctx context.Context, c *Container, volumeSet map[string]struct{}) error {
	var err error

	logrus.Debugf("bind volumes: %v", c.HostConfig.Binds)

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
			return errors.Errorf("unknown bind: %s", b)
		}

		if opts.CheckDuplicateMountPoint(c.Mounts, mp.Destination) {
			logrus.Warnf("duplicate mount point: %s", mp.Destination)
			continue
		}

		if mp.Source == "" {
			mp.Source = randomid.Generate()
		}

		err = opts.ParseBindMode(mp, mode)
		if err != nil {
			logrus.Errorf("failed to parse bind mode(%s): %v", mode, err)
			return err
		}

		if !path.IsAbs(mp.Source) {
			// volume bind.
			name := mp.Source
			if _, exist := volumeSet[name]; !exist {
				mp.Name = name
				mp.Source, mp.Driver, err = mgr.attachVolume(ctx, name, c)
				if err != nil {
					logrus.Errorf("failed to bind volume(%s): %v", name, err)
					return errors.Wrap(err, "failed to bind volume")
				}

				volumeSet[mp.Name] = struct{}{}
			}

			if mp.Replace != "" {
				mp.Source, err = mgr.VolumeMgr.Path(ctx, name)
				if err != nil {
					return err
				}

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
		}

		if _, err = os.Stat(mp.Source); err != nil {
			// host directory bind into container.
			if !os.IsNotExist(err) {
				return errors.Errorf("failed to stat %q: %v", mp.Source, err)
			}
			// Create the host path if it doesn't exist.
			if err = os.MkdirAll(mp.Source, 0755); err != nil {
				return errors.Errorf("failed to mkdir %q: %v", mp.Source, err)
			}
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
			logrus.Warnf("duplicate mount point: %s from volumes", dest)
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
			logrus.Errorf("failed to bind volume(%s): %v", mp.Name, err)
			return errors.Wrap(err, "failed to bind volume")
		}

		err = opts.ParseBindMode(mp, "")
		if err != nil {
			logrus.Errorf("failed to parse mode: %v", err)
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
		return errors.Wrapf(err, "failed to get image: %s", c.Image)
	}
	for dest := range image.Config.Volumes {
		// check if volume has been created
		name := randomid.Generate()
		if _, exist := volumeSet[name]; exist {
			continue
		}

		if opts.CheckDuplicateMountPoint(c.Mounts, dest) {
			logrus.Warnf("duplicate mount point: %s from image", dest)
			continue
		}

		mp := new(types.MountPoint)
		mp.Name = name
		mp.Destination = dest

		mp.Source, mp.Driver, err = mgr.attachVolume(ctx, mp.Name, c)
		if err != nil {
			logrus.Errorf("failed to bind volume(%s): %v", mp.Name, err)
			return errors.Wrap(err, "failed to bind volume")
		}

		err = opts.ParseBindMode(mp, "")
		if err != nil {
			logrus.Errorf("failed to parse mode: %v", err)
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
				RW:          oldMountPoint.RW,
				Propagation: oldMountPoint.Propagation,
			}

			if _, exist := volumeSet[oldMountPoint.Name]; len(oldMountPoint.Name) > 0 && !exist {
				mp.Name = oldMountPoint.Name
				mp.Source, mp.Driver, err = mgr.attachVolume(ctx, oldMountPoint.Name, container)
				if err != nil {
					logrus.Errorf("failed to bind volume(%s): %v", oldMountPoint.Name, err)
					return errors.Wrap(err, "failed to bind volume")
				}

				volumeSet[mp.Name] = struct{}{}
			}

			err = opts.ParseBindMode(mp, mode)
			if err != nil {
				logrus.Errorf("failed to parse volumes-from mode %s: %v", mode, err)
				return err
			}

			// the volumes from VolumeFrom is not allowed to CopyData
			mp.CopyData = false

			container.Mounts = append(container.Mounts, mp)
		}
	}

	return nil
}

func (mgr *ContainerManager) populateVolumes(ctx context.Context, c *Container) error {
	for _, mnt := range c.Mounts {
		if mnt.Driver == "tmpfs" {
			continue
		}

		if !mnt.CopyData {
			continue
		}

		logrus.Debugf("copying image data from %s:%s, to %s, path: %s",
			c.ID, mnt.Destination, mnt.Name, mnt.Source)

		imagePath := path.Join(c.MountFS, mnt.Destination)

		err := copyImageContent(imagePath, mnt.Source)
		if err != nil {
			logrus.Errorf("failed to populate volume[name: %s, source: %s]: %v", mnt.Name, mnt.Source, err)
			return err
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
		return errors.Wrapf(err, "failed to write file: (%s)", mtabPath)
	}

	logrus.Infof("write mtab file to (%s)", mtabPath)

	return nil
}

func (mgr *ContainerManager) setRootfsQuota(ctx context.Context, c *Container) error {
	logrus.Debugf("start to set rootfs quota, directory: (%s)", c.MountFS)

	if c.MountFS == "" {
		return nil
	}

	rootfsQuota := quota.GetDefaultQuota(c.Config.DiskQuota)
	if rootfsQuota == "" {
		return nil
	}

	qid := "0"
	if c.Config.QuotaID != "" {
		qid = c.Config.QuotaID
	}

	id, err := strconv.Atoi(qid)
	if err != nil {
		return errors.Wrapf(err, "failed to change quota id: (%s) from string to int", qid)
	}

	// set rootfs quota
	_, err = quota.SetRootfsDiskQuota(c.MountFS, rootfsQuota, uint32(id))
	if err != nil {
		return errors.Wrapf(err, "failed to set rootfs quota, mountfs: (%s), quota: (%s), quota id: (%d)",
			c.MountFS, rootfsQuota, id)
	}

	return nil
}

func (mgr *ContainerManager) setMountPointDiskQuota(ctx context.Context, c *Container) error {
	if c.Config.DiskQuota == nil {
		if c.Config.QuotaID != "" && c.Config.QuotaID != "0" {
			return fmt.Errorf("invalid argument, set quota-id without disk-quota")
		}
		return nil
	}

	var (
		qid        uint32
		setQuotaID bool
	)

	if c.Config.QuotaID != "" {
		id, err := strconv.Atoi(c.Config.QuotaID)
		if err != nil {
			return errors.Wrapf(err, "invalid argument, QuotaID: %s", c.Config.QuotaID)
		}

		// if QuotaID is < 0, it means pouchd alloc a unique quota id.
		if id < 0 {
			qid, err = quota.GetNextQuotaID()
			if err != nil {
				return errors.Wrap(err, "failed to get next quota id")
			}

			// update QuotaID
			c.Config.QuotaID = strconv.Itoa(int(qid))
		} else {
			qid = uint32(id)
		}
	}

	if qid > 0 {
		setQuotaID = true
	}

	// get rootfs quota
	quotas := c.Config.DiskQuota
	defaultQuota := quota.GetDefaultQuota(quotas)
	if setQuotaID && defaultQuota == "" {
		return fmt.Errorf("set quota id but have no set default quota size")
	}

	// parse diskquota regexe
	var res []*quota.RegExp
	for path, size := range quotas {
		re := regexp.MustCompile(path)
		res = append(res, &quota.RegExp{Pattern: re, Path: path, Size: size})
	}

	for _, mp := range c.Mounts {
		// skip volume mount or replace mode mount
		if mp.Replace != "" || mp.Source == "" || mp.Destination == "" {
			logrus.Debugf("skip volume mount or replace mode mount")
			continue
		}

		if mp.Name != "" {
			v, err := mgr.VolumeMgr.Get(ctx, mp.Name)
			if err != nil {
				logrus.Warnf("failed to get volume: %s", mp.Name)
				continue
			}

			if v.Size() != "" {
				logrus.Debugf("skip volume: %s with size", mp.Name)
				continue
			}
		}

		// skip non-directory path.
		if fd, err := os.Stat(mp.Source); err != nil || !fd.IsDir() {
			logrus.Debugf("skip non-directory path: %s", mp.Source)
			continue
		}

		matched := false
		for _, re := range res {
			findStr := re.Pattern.FindString(mp.Destination)
			if findStr == mp.Destination {
				quotas[mp.Destination] = re.Size
				matched = true
				if re.Path != ".*" {
					break
				}
			}
		}

		size := ""
		if matched && !setQuotaID {
			size = quotas[mp.Destination]
		} else {
			size = defaultQuota
		}
		err := quota.SetDiskQuota(mp.Source, size, qid)
		if err != nil {
			// just ignore set disk quota fail
			logrus.Warnf("failed to set disk quota, directory: (%s), size: (%s), quota id: (%d), err: (%v)",
				mp.Source, size, qid, err)
		}
	}

	c.Config.DiskQuota = quotas

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
			logrus.Warnf("failed to detach volume: %s, err: %v", name, err)
		}

		if remove && !mount.Named {
			if err := mgr.VolumeMgr.Remove(ctx, name); err != nil && !errtypes.IsInUse(err) {
				logrus.Warnf("failed to remove volume: %s when remove container", name)
			}
		}
	}

	return nil
}

// setMountFS is used to set mountfs directory.
func (mgr *ContainerManager) setMountFS(ctx context.Context, c *Container) {
	c.Lock()
	defer c.Unlock()

	c.MountFS = path.Join(mgr.Store.Path(c.ID), "rootfs")
}

// Mount sets the container rootfs
func (mgr *ContainerManager) Mount(ctx context.Context, c *Container) error {
	if c.MountFS == "" {
		mgr.setMountFS(ctx, c)
	}

	mounts, err := mgr.Client.GetMounts(ctx, c.ID)
	if err != nil {
		return err
	} else if len(mounts) != 1 {
		return fmt.Errorf("failed to get snapshot %s mounts: not equals 1", c.ID)
	}

	err = os.MkdirAll(c.MountFS, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return mounts[0].Mount(c.MountFS)
}

// Unmount unsets the container rootfs
func (mgr *ContainerManager) Unmount(ctx context.Context, c *Container) error {
	// TODO: if umount is failed, and how to deal it.
	err := mount.Unmount(c.MountFS, 0)
	if err != nil {
		return errors.Wrapf(err, "failed to umount mountfs: (%s)", c.MountFS)
	}

	return os.RemoveAll(c.MountFS)
}

func (mgr *ContainerManager) initContainerStorage(ctx context.Context, c *Container) (err error) {
	if err = mgr.Mount(ctx, c); err != nil {
		return errors.Wrapf(err, "failed to mount rootfs: (%s)", c.MountFS)
	}

	defer func() {
		if umountErr := mgr.Unmount(ctx, c); umountErr != nil {
			if err != nil {
				err = errors.Wrapf(err, "failed to umount rootfs: (%s), err: (%v)", c.MountFS, umountErr)
			} else {
				err = errors.Wrapf(umountErr, "failed to umount rootfs: (%s)", c.MountFS)
			}
		}
	}()

	// try to setup container working directory
	if err := mgr.SetupWorkingDirectory(ctx, c); err != nil {
		return errors.Wrapf(err, "failed to setup container %s working directory", c.ID)
	}

	// parse volume config
	if err = mgr.generateMountPoints(ctx, c); err != nil {
		return errors.Wrap(err, "failed to parse volume argument")
	}

	// set mount point disk quota
	if err = mgr.setMountPointDiskQuota(ctx, c); err != nil {
		return errors.Wrap(err, "failed to set mount point disk quota")
	}

	// set rootfs disk quota
	if err = mgr.setRootfsQuota(ctx, c); err != nil {
		logrus.Warnf("failed to set rootfs disk quota, err: %v", err)
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

	c.Config.WorkingDir = filepath.Clean(c.Config.WorkingDir)

	path := filepath.Join(c.MountFS, c.Config.WorkingDir)
	// TODO(ziren): not care about File mode
	err := os.MkdirAll(path, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
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

	volList, err := ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	if len(volList) > 0 {
		fi, err := os.Stat(destination)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		} else if !fi.IsDir() {
			return nil
		}

		dstList, err := ioutil.ReadDir(destination)
		if err != nil {
			return err
		}
		if len(dstList) == 0 {
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
	return os.Chmod(destination, os.FileMode(fi.Mode()))
}
