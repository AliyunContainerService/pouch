// +build linux

package quota

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/alibaba/pouch/pkg/bytefmt"
	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/log"
	"github.com/alibaba/pouch/pkg/system"

	"github.com/pkg/errors"
)

// GrpQuotaDriver represents group quota driver.
type GrpQuotaDriver struct {
	lock sync.Mutex

	// quotaIDs saves all of quota ids.
	// key: quota ID which means this ID is used in the global scope.
	// value: stuct{}
	quotaIDs map[uint32]struct{}

	// mountPoints saves all the mount point of volume which have already been enforced disk quota.
	// key: device ID such as /dev/sda1
	// value: the mountpoint of the device in the filesystem
	mountPoints map[uint64]string

	// LastID is used to mark last used quota ID.
	// quota ID is allocated increasingly by sequence one by one.
	lastID uint32
}

// EnforceQuota is used to enforce disk quota effect on specified directory.
func (quota *GrpQuotaDriver) EnforceQuota(dir string) (string, error) {
	log.With(nil).Debugf("start group quota driver: (%s)", dir)

	devID, err := system.GetDevID(dir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get deivce id for directory: (%s)", dir)
	}

	// set limit of dir's device in driver
	if _, err = setDevLimit(dir, devID); err != nil {
		return "", errors.Wrapf(err, "failed to set device limit, dir: (%s), devID: (%d)", dir, devID)
	}

	quota.lock.Lock()
	defer quota.lock.Unlock()

	if mp, ok := quota.mountPoints[devID]; ok {
		// if the device has already been enforced quota, just return.
		return mp, nil
	}

	mountPoint, hasQuota, _ := quota.CheckMountpoint(devID)
	if len(mountPoint) == 0 {
		return mountPoint, fmt.Errorf("failed to find mountpoint: (%s)", dir)
	}
	if !hasQuota {
		// remount option grpquota for mountpoint
		exit, stdout, stderr, err := exec.Run(0, "mount", "-o", "remount,grpquota", mountPoint)
		if err != nil {
			log.With(nil).Errorf("failed to remount grpquota, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
				mountPoint, stdout, stderr, exit, err)
			return "", errors.Wrapf(err, "failed to remount grpquota, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
				mountPoint, stdout, stderr, exit)
		}
	}

	vfsVersion, quotaFilename, err := getVFSVersionAndQuotaFile(devID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get vfs version and quota file")
	}

	filename := mountPoint + "/" + quotaFilename
	if _, err := os.Stat(filename); err != nil {
		os.Remove(mountPoint + "/aquota.user")

		header := []byte{0x27, 0x19, 0xc0, 0xd9, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3a, 0x09, 0x00, 0x80,
			0x3a, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x05, 0x00, 0x00, 0x00}
		if vfsVersion == "vfsv1" {
			header[4] = 0x01
		}

		if writeErr := ioutil.WriteFile(filename, header, 0644); writeErr != nil {
			return mountPoint, errors.Wrapf(writeErr, "failed to write file, filename: (%s), vfs version: (%s)",
				filename, vfsVersion)
		}
		if exit, stdout, stderr, err := exec.Run(0, "setquota", "-g", "-t", "43200", "43200", mountPoint); err != nil {
			os.Remove(filename)
			log.With(nil).Errorf("failed to setquota, stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
				stdout, stderr, exit, err)
			return mountPoint, errors.Wrapf(err, "failed to setquota, stdout: (%s), stderr: (%s), exit: (%d)",
				stdout, stderr, exit)
		}
		if err := quota.setQuota(0, 0, mountPoint); err != nil {
			os.Remove(filename)
			log.With(nil).Errorf("failed to set quota, mountpoint: (%s), err: (%v)", mountPoint, err)
			return mountPoint, errors.Wrapf(err, "failed to set quota, mountpoint: (%s)", mountPoint)
		}
	}

	// check group quota status, on or not, pay attention, the right exit code of command 'quotaon' is '1'.
	exit, stdout, stderr, err := exec.Run(0, "quotaon", "-pg", mountPoint)
	if err != nil && exit != 1 {
		log.With(nil).Errorf("failed to quota on for mountpoint: (%s), exit: (%d), stdout: (%s), stderr: (%s), err: (%v)",
			mountPoint, exit, stdout, stderr, err)
		return "", errors.Wrapf(err, "failed to quota on for mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
			mountPoint, stdout, stderr, exit)
	}
	if strings.Contains(stdout, " is on") {
		quota.mountPoints[devID] = mountPoint
		return mountPoint, nil
	}
	if exit, stdout, stderr, err = exec.Run(0, "quotaon", mountPoint); err != nil {
		mountPoint = ""
		err = errors.Wrapf(err, "failed to quotaon, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
			mountPoint, stdout, stderr, exit)
	}

	quota.mountPoints[devID] = mountPoint
	return mountPoint, err
}

// SetSubtree is used to set quota id for directory,
// setfattr -n system.subtree -v $QUOTAID
func (quota *GrpQuotaDriver) SetSubtree(dir string, qid uint32) (uint32, error) {
	log.With(nil).Debugf("set subtree, dir: %s, quotaID: %d", dir, qid)
	id := qid
	var err error
	if id == 0 {
		id = quota.GetQuotaIDInFileAttr(dir)
		if id > 0 {
			return id, nil
		}
		id, err = quota.GetNextQuotaID()
	}

	if err != nil {
		return 0, errors.Wrapf(err, "failed to get file: (%s) quota id", dir)
	}
	strid := strconv.FormatUint(uint64(id), 10)
	exit, stdout, stderr, err := exec.Run(0, "setfattr", "-n", "system.subtree", "-v", strid, dir)

	return id, errors.Wrapf(err, "failed to setfattr, dir: (%s), quota id: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
		dir, strid, stdout, stderr, exit)
}

// CheckMountpoint is used to check mount point.
// It returns mointpoint, enable quota and filesystem type of the device.
//
// cat /proc/mounts as follows:
// /dev/sda3 / ext4 rw,relatime,data=ordered 0 0
// /dev/sda2 /boot/grub2 ext4 rw,relatime,stripe=4,data=ordered 0 0
// /dev/sda5 /home ext4 rw,relatime,data=ordered 0 0
// /dev/sdb1 /home/pouch ext4 rw,relatime,prjquota,data=ordered 0 0
// tmpfs /run tmpfs rw,nosuid,nodev,mode=755 0 0
// tmpfs /sys/fs/cgroup tmpfs ro,nosuid,nodev,noexec,mode=755 0 0
// cgroup /sys/fs/cgroup/cpuset,cpu,cpuacct cgroup rw,nosuid,nodev,noexec,relatime,cpuacct,cpu,cpuset 0 0
// cgroup /sys/fs/cgroup/devices cgroup rw,nosuid,nodev,noexec,relatime,devices 0 0
// cgroup /sys/fs/cgroup/memory cgroup rw,nosuid,nodev,noexec,relatime,memory 0 0
// cgroup /sys/fs/cgroup/blkio cgroup rw,nosuid,nodev,noexec,relatime,blkio 0 0
func (quota *GrpQuotaDriver) CheckMountpoint(devID uint64) (string, bool, string) {
	log.With(nil).Debugf("check mountpoint, devID: %d", devID)
	output, err := ioutil.ReadFile(procMountFile)
	if err != nil {
		log.With(nil).Warnf("failed to read file: (%s), err: (%v)", procMountFile, err)
		return "", false, ""
	}

	var (
		enableQuota bool
		mountPoint  string
		fsType      string
	)

	// Two formats of group quota.
	// /dev/sdb1 /home/pouch ext4 rw,relatime,prjquota,data=ordered 0 0
	// /dev/sda1 /home/pouch ext4 rw,relatime,data=ordered,jqfmt=vfsv0,grpjquota=aquota.group 0 0
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) != 6 {
			continue
		}

		devID2, _ := system.GetDevID(parts[1])
		if devID != devID2 {
			continue
		}

		// check the shortest mountpoint.
		if mountPoint != "" && len(mountPoint) < len(parts[1]) {
			continue
		}

		// get device's mountpoint and fs type.
		mountPoint = parts[1]
		fsType = parts[2]

		// check the device turn on the grpquota or not.
		if strings.Contains(parts[3], "grpquota") || strings.Contains(parts[3], "grpjquota") {
			enableQuota = true
		}
	}

	log.With(nil).Debugf("check device: (%d), mountpoint: (%s), enableQuota: (%v), fsType: (%s)",
		devID, mountPoint, enableQuota, fsType)

	return mountPoint, enableQuota, fsType
}

// SetDiskQuota is used to set quota for directory.
func (quota *GrpQuotaDriver) SetDiskQuota(dir string, size string, quotaID uint32) error {
	log.With(nil).Debugf("set disk quota, dir: %s, size: %s, quotaID: %d", dir, size, quotaID)

	mountPoint, err := quota.EnforceQuota(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to enforce quota, dir: (%s)", dir)
	}
	if len(mountPoint) == 0 {
		return errors.Errorf("failed to find mountpoint, dir: (%s)", dir)
	}

	// transfer limit from kbyte to byte
	limit, err := bytefmt.ToKilobytes(size)
	if err != nil {
		return errors.Wrapf(err, "failed to change size: (%s) to kilobytes", size)
	}

	if err := checkDevLimit(dir, limit*1024); err != nil {
		return err
	}

	id, err := quota.SetSubtree(dir, quotaID)
	if err != nil {
		return errors.Wrapf(err, "failed to set subtree, dir: (%s), quota id: (%d)", dir, quotaID)
	}
	if id == 0 {
		return errors.Errorf("failed to find quota id to set subtree")
	}

	return quota.setQuota(id, limit, mountPoint)
}

func (quota *GrpQuotaDriver) setQuota(quotaID uint32, diskQuota uint64, mountPoint string) error {
	log.With(nil).Debugf("set user quota, quotaID: %d, limit: %d, mountpoint: %s", quotaID, diskQuota, mountPoint)

	quotaIDStr := strconv.FormatUint(uint64(quotaID), 10)
	limit := strconv.FormatUint(diskQuota, 10)

	exit, stdout, stderr, err := exec.Run(0, "setquota", "-g", quotaIDStr, "0", limit, "0", "0", mountPoint)
	return errors.Wrapf(err, "failed to set quota, mountpoint: (%s), quota id: (%d), quota: (%d kbytes), stdout: (%s), stderr: (%s), exit: (%d)",
		mountPoint, quotaID, diskQuota, stdout, stderr, exit)
}

// GetQuotaIDInFileAttr returns quota ID in the directory attributes.
// getfattr -n system.subtree --only-values --absolute-names /
func (quota *GrpQuotaDriver) GetQuotaIDInFileAttr(dir string) uint32 {
	log.With(nil).Debugf("get file attr, dir: %s", dir)

	exit, stdout, stderr, err := exec.Run(0, "getfattr", "-n", "system.subtree", "--only-values", "--absolute-names", dir)
	if err != nil {
		log.With(nil).Errorf("failed to getfattr, dir: (%s), stdout: (%s), stderr: (%s), exit: (%d), err: (%s)",
			dir, stdout, stderr, exit, err)
		return 0
	}
	v, _ := strconv.Atoi(stdout)
	return uint32(v)
}

// SetQuotaIDInFileAttr is used to set quota ID in file attributes.
func (quota *GrpQuotaDriver) SetQuotaIDInFileAttr(dir string, id uint32) error {
	log.With(nil).Debugf("set file attr, dir: %s, quotaID: %d", dir, id)

	strid := strconv.FormatUint(uint64(id), 10)
	exit, stdout, stderr, err := exec.Run(0, "setfattr", "-n", "system.subtree", "-v", strid, dir)
	return errors.Wrapf(err, "failed to setfattr, dir: (%s), quota id: (%d), stdout: (%s), stderr: (%s), exit: (%d)",
		dir, id, stdout, stderr, exit)
}

// SetQuotaIDInFileAttrNoOutput is used to set file attributes without error.
func (quota *GrpQuotaDriver) SetQuotaIDInFileAttrNoOutput(dir string, quotaID uint32) {
	strid := strconv.FormatUint(uint64(quotaID), 10)
	exit, stdout, stderr, err := exec.Run(0, "setfattr", "-n", "system.subtree", "-v", strid, dir)
	if err != nil {
		log.With(nil).Errorf("failed to setfattr, dir: (%s), quota id: (%d), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
			dir, quotaID, stdout, stderr, exit, err)
	}
}

// GetNextQuotaID returns the next available quota id.
func (quota *GrpQuotaDriver) GetNextQuotaID() (uint32, error) {
	quota.lock.Lock()
	defer quota.lock.Unlock()

	if quota.lastID == 0 {
		var err error
		quota.quotaIDs, quota.lastID, err = loadQuotaIDs("-gan")
		if err != nil {
			return 0, errors.Wrap(err, "failed to load quota list")
		}
	}
	id := quota.lastID
	for {
		if id < QuotaMinID {
			id = QuotaMinID
		}
		id++
		if _, ok := quota.quotaIDs[id]; !ok {
			break
		}
	}
	quota.quotaIDs[id] = struct{}{}
	quota.lastID = id

	log.With(nil).Debugf("get next project quota id: %d", id)
	return id, nil
}

func getVFSVersionAndQuotaFile(devID uint64) (string, string, error) {
	output, err := ioutil.ReadFile(procMountFile)
	if err != nil {
		log.With(nil).Warnf("failed to read file: (%s), err: (%v)", procMountFile, err)
		return "", "", errors.Wrap(err, "failed to read /proc/mounts")
	}

	vfsVersion := "vfsv0"
	quotaFilename := "aquota.group"
	for _, line := range strings.Split(string(output), "\n") {
		// TODO: add an example here to make following code readable.
		// /dev/sdb1 /home/pouch ext4 rw,relatime,prjquota,data=ordered 0 0 ?
		parts := strings.Split(line, " ")
		if len(parts) != 6 {
			continue
		}

		devID2, _ := system.GetDevID(parts[1])
		if devID != devID2 {
			continue
		}

		for _, opt := range strings.Split(parts[3], ",") {
			items := strings.SplitN(opt, "=", 2)
			if len(items) != 2 {
				continue
			}
			switch items[0] {
			case "jqfmt":
				vfsVersion = items[1]
			case "grpjquota":
				quotaFilename = items[1]
			}
		}
		return vfsVersion, quotaFilename, nil
	}

	return vfsVersion, quotaFilename, nil
}

func (quota *GrpQuotaDriver) SetFileAttrRecursive(dir string, quotaID uint32) error {
	return filepath.Walk(dir, func(path string, fd os.FileInfo, err error) error {
		if err != nil {
			log.With(nil).Warnf("setQuota walk dir %s get error %v", path, err)
			return nil
		}

		existedQid := quota.GetQuotaIDInFileAttr(path)
		if existedQid != quotaID {
			quota.SetQuotaIDInFileAttrNoOutput(path, quotaID)
		}
		return nil
	})
}
