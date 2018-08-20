// +build linux

package quota

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/alibaba/pouch/pkg/bytefmt"
	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/system"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

var (
	prjQuotaType = "prjquota"
)

// PrjQuotaDriver represents project quota driver.
type PrjQuotaDriver struct {
	lock sync.Mutex

	// quotaIDs saves all of quota ids.
	// key: quota ID which means this ID is used in the global scope.
	// value: stuct{}
	quotaIDs map[uint32]struct{}

	// mountPoints saves all the mount point of volume which have already been enforced disk quota.
	// key: device ID such as /dev/sda1
	// value: the mountpoint of the device in the filesystem
	mountPoints map[uint64]string

	// devLimits saves all the limit of device.
	// key: device ID
	// value: the storage upper limit size of the device(unit:B)
	devLimits map[uint64]uint64

	// lastID is used to mark last used quota ID.
	// quota ID is allocated increasingly by sequence one by one.
	lastID uint32
}

// EnforceQuota is used to enforce disk quota effect on specified directory.
func (quota *PrjQuotaDriver) EnforceQuota(dir string) (string, error) {
	logrus.Debugf("start project quota driver: (%s)", dir)

	devID, err := system.GetDevID(dir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get device id for directory: (%s)", dir)
	}

	// set limit of dir's device in driver
	if _, err = quota.setDevLimit(dir, devID); err != nil {
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
		return mountPoint, fmt.Errorf("mountPoint not found for the device on which dir (%s) lies", dir)
	}
	if !hasQuota {
		// remount option prjquota for mountpoint
		exit, stdout, stderr, err := exec.Run(0, "mount", "-o", "remount,prjquota", mountPoint)
		if err != nil {
			logrus.Errorf("failed to remount prjquota, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
				mountPoint, stdout, stderr, exit, err)
			return "", errors.Wrapf(err, "failed to remount prjquota, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
				mountPoint, stdout, stderr, exit)
		}
	}

	// use tool quotaon to set disk quota for mountpoint
	exit, stdout, stderr, err := exec.Run(0, "quotaon", "-P", mountPoint)
	if err != nil {
		if strings.Contains(stderr, " File exists") {
			err = nil
		} else {
			logrus.Errorf("failed to quota on, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
				mountPoint, stdout, stderr, exit, err)
			err = errors.Wrapf(err, "failed to quota on, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
				mountPoint, stdout, stderr, exit)
			mountPoint = ""
		}
	}

	// record device which has quota settings
	quota.mountPoints[devID] = mountPoint

	return mountPoint, err
}

// SetSubtree is used to set quota id for substree dir which is container's root dir.
// For container, it has its own root dir.
// And this dir is a subtree of the host dir which is mapped to a device.
// chattr -p qid +P $QUOTAID
func (quota *PrjQuotaDriver) SetSubtree(dir string, qid uint32) (uint32, error) {
	logrus.Debugf("set subtree, dir: %s, quotaID: %d", dir, qid)
	id := qid
	var err error

	if id == 0 {
		id = quota.GetQuotaIDInFileAttr(dir)
		if id > 0 {
			return id, nil
		}
		if id, err = quota.GetNextQuotaID(); err != nil {
			return 0, errors.Wrapf(err, "failed to get file: (%s) quota id", dir)
		}
	}

	strid := strconv.FormatUint(uint64(id), 10)
	exit, stdout, stderr, err := exec.Run(0, "chattr", "-p", strid, "+P", dir)
	return id, errors.Wrapf(err, "failed to chattr, dir: (%s), quota id: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
		dir, strid, stdout, stderr, exit)
}

// SetDiskQuota uses the following two parameters to set disk quota for a directory.
// * quota size: a byte size of requested quota.
// * quota ID: an ID represent quota attr which is used in the global scope.
func (quota *PrjQuotaDriver) SetDiskQuota(dir string, size string, quotaID uint32) error {
	logrus.Debugf("set disk quota, dir: %s, size: %s, quotaID: %d", dir, size, quotaID)
	mountPoint, err := quota.EnforceQuota(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to enforce quota, dir: (%s)", dir)
	}
	if len(mountPoint) == 0 {
		return errors.Errorf("failed to find mountpoint, dir: (%s)", dir)
	}

	id, err := quota.SetSubtree(dir, quotaID)
	if err != nil {
		return errors.Wrapf(err, "failed to set subtree, dir: (%s), quota id: (%d)", dir, quotaID)
	}
	if id == 0 {
		return errors.Errorf("failed to find quota id to set subtree")
	}

	limit, err := bytefmt.ToKilobytes(size)
	if err != nil {
		return errors.Wrapf(err, "failed to change size: (%s) to kilobytes", size)
	}

	// transfer limit from kbyte to byte
	if err := quota.checkDevLimit(dir, limit*1024); err != nil {
		return errors.Wrapf(err, "failed to check device limit, dir: (%s), limit: (%d)kb", dir, limit)
	}

	return quota.setQuota(id, limit, mountPoint)
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
func (quota *PrjQuotaDriver) CheckMountpoint(devID uint64) (string, bool, string) {
	logrus.Debugf("check mountpoint, devID: %d", devID)
	output, err := ioutil.ReadFile(procMountFile)
	if err != nil {
		logrus.Warnf("failed to read file: (%s), err: (%v)", procMountFile, err)
		return "", false, ""
	}

	// /dev/sdb1 /home/pouch ext4 rw,relatime,prjquota,data=ordered 0 0
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) != 6 {
			continue
		}

		mountPoint := parts[1]
		fsType := parts[2]

		devID2, _ := system.GetDevID(mountPoint)
		if devID != devID2 {
			continue
		}

		for _, value := range strings.Split(parts[3], ",") {
			if value == "prjquota" {
				return mountPoint, true, fsType
			}
		}
	}

	return "", false, ""
}

// setQuota uses system tool "setquota" to set project quota for binding of limit and mountpoint and quotaID.
// * quotaID: quota ID which means this ID is used in the global scope.
// * blockLimit: block limit number for mountpoint.
// * mountPoint: the mountpoint of the device in the filesystem
func (quota *PrjQuotaDriver) setQuota(quotaID uint32, blockLimit uint64, mountPoint string) error {
	logrus.Debugf("set project quota, quotaID: %d, limit: %d, mountpoint: %s", quotaID, blockLimit, mountPoint)

	quotaIDStr := strconv.FormatUint(uint64(quotaID), 10)
	blockLimitStr := strconv.FormatUint(blockLimit, 10)
	// set project quota
	exit, stdout, stderr, err := exec.Run(0, "setquota", "-P", quotaIDStr, "0", blockLimitStr, "0", "0", mountPoint)
	return errors.Wrapf(err, "failed to set quota, mountpoint: (%s), quota id: (%d), quota: (%d kbytes), stdout: (%s), stderr: (%s), exit: (%d)",
		mountPoint, quotaID, blockLimit, stdout, stderr, exit)
}

// GetQuotaIDInFileAttr gets attributes of the file which is in the inode.
// The returned result is quota ID.
// return 0 if failure happens, since quota ID must be positive.
// execution command: `lsattr -p $dir`
func (quota *PrjQuotaDriver) GetQuotaIDInFileAttr(dir string) uint32 {
	parent := path.Dir(dir)
	qid := 0

	exit, stdout, stderr, err := exec.Run(0, "lsattr", "-p", parent)
	if err != nil {
		// failure, then return invalid value 0 for quota ID.
		logrus.Errorf("failed to lsattr, dir: (%s), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
			dir, stdout, stderr, exit, err)
		return 0
	}

	// example output:
	// 16777256 --------------e---P ./exampleDir
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) > 2 && parts[2] == dir {
			// find the corresponding quota ID, return directly.
			qid, _ = strconv.Atoi(parts[0])
			logrus.Debugf("get file attr: [%s], quota id: [%d]", dir, qid)
			return uint32(qid)
		}
	}

	logrus.Errorf("failed to get file attr of quota ID for dir %s", dir)
	return 0
}

// SetQuotaIDInFileAttr sets file attributes of quota ID for the input directory.
// The input attributes is quota ID.
func (quota *PrjQuotaDriver) SetQuotaIDInFileAttr(dir string, quotaID uint32) error {
	logrus.Debugf("set file attr, dir: %s, quotaID: %d", dir, quotaID)

	strid := strconv.FormatUint(uint64(quotaID), 10)
	exit, stdout, stderr, err := exec.Run(0, "chattr", "-p", strid, "+P", dir)
	return errors.Wrapf(err, "failed to chattr, dir: (%s), quota id: (%d), stdout: (%s), stderr: (%s), exit: (%d)",
		dir, quotaID, stdout, stderr, exit)
}

// SetQuotaIDInFileAttrNoOutput is used to set file attributes without error.
func (quota *PrjQuotaDriver) SetQuotaIDInFileAttrNoOutput(dir string, quotaID uint32) {
	strid := strconv.FormatUint(uint64(quotaID), 10)
	exit, stdout, stderr, err := exec.Run(0, "chattr", "-p", strid, "+P", dir)
	if err != nil {
		logrus.Errorf("failed to chattr, dir: (%s), quota id: (%d), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
			dir, quotaID, stdout, stderr, exit, err)
	}
}

// GetNextQuotaID returns the next available quota id.
func (quota *PrjQuotaDriver) GetNextQuotaID() (uint32, error) {
	quota.lock.Lock()
	defer quota.lock.Unlock()

	if quota.lastID == 0 {
		var err error
		quota.quotaIDs, quota.lastID, err = loadQuotaIDs("-Pan")
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

	logrus.Debugf("get next project quota id: %d", id)
	return id, nil
}

// setDevLimit sets device storage upper limit in quota driver according to inpur dir.
func (quota *PrjQuotaDriver) setDevLimit(dir string, devID uint64) (uint64, error) {
	if limit, exist := quota.devLimits[devID]; exist {
		return limit, nil
	}

	// get storage upper limit of the device which the dir is on.
	var stfs syscall.Statfs_t
	if err := syscall.Statfs(dir, &stfs); err != nil {
		logrus.Errorf("failed to get path: (%s) limit, err: (%v)", dir, err)
		return 0, errors.Wrapf(err, "failed to get path: (%s) limit", dir)
	}
	limit := stfs.Blocks * uint64(stfs.Bsize)

	quota.lock.Lock()
	quota.devLimits[devID] = limit
	quota.lock.Unlock()

	logrus.Debugf("SetDevLimit: dir %s limit is %v B", dir, limit)
	return limit, nil
}

// checkDevLimit checks if the device on which the input dir lies has already been recorded in driver.
func (quota *PrjQuotaDriver) checkDevLimit(dir string, size uint64) error {
	devID, err := system.GetDevID(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to get device id, dir: (%s)", dir)
	}

	limit, exist := quota.devLimits[devID]
	if !exist {
		// if has not recorded, just add (dir, device, limit) to driver.
		if limit, err = quota.setDevLimit(dir, devID); err != nil {
			return errors.Wrapf(err, "failed to set device limit, dir: (%s), devID: (%d)", dir, devID)
		}
	}

	if limit < size {
		return fmt.Errorf("dir %s quota limit %v must be less than %v", dir, size, limit)
	}

	logrus.Debugf("succeeded in checkDevLimit (dir %s quota limit %v B) with size %v B", dir, limit, size)

	return nil
}
