// +build linux

package quota

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/alibaba/pouch/pkg/bytefmt"
	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/log"
	"github.com/alibaba/pouch/pkg/system"

	"github.com/pkg/errors"
)

// PrjQuotaDriver represents project quota driver.
type PrjQuotaDriver struct {
	lock sync.Mutex

	// quotaIDs saves all of quota ids.
	// key: quota ID which means this ID is used in the global scope.
	// value: stuct{}
	quotaIDs map[uint32]struct{}

	// lastID is used to mark last used quota ID.
	// quota ID is allocated increasingly by sequence one by one.
	lastID uint32
}

// EnforceQuota is used to enforce disk quota effect on specified directory.
// it returns the mountpoint and error.
func (quota *PrjQuotaDriver) EnforceQuota(dir string) (*MountInfo, error) {
	log.With(nil).Debugf("start project quota driver: (%s)", dir)

	// get device id for set directory.
	devID, err := getDevID(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get device id for directory: (%s)", dir)
	}

	mountPoint, hasQuota, fsType := quota.CheckMountpoint(devID)
	if mountPoint == "" {
		return nil, fmt.Errorf("mountPoint not found for the device on which dir (%s) lies", dir)
	}
	if !hasQuota {
		// remount option prjquota for mountpoint
		exit, stdout, stderr, err := exec.Run(0, "mount", "-o", "remount,prjquota", mountPoint)
		if err != nil {
			log.With(nil).Errorf("failed to remount prjquota, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
				mountPoint, stdout, stderr, exit, err)
			return nil, errors.Wrapf(err, "failed to remount prjquota, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
				mountPoint, stdout, stderr, exit)
		}
	}

	// use tool quotaon to set disk quota for mountpoint
	exit, stdout, stderr, err := exec.Run(0, "quotaon", "-P", mountPoint)
	if err != nil {
		if strings.Contains(stderr, " File exists") {
			err = nil
		} else {
			log.With(nil).Errorf("failed to quota on, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
				mountPoint, stdout, stderr, exit, err)
			err = errors.Wrapf(err, "failed to quota on, mountpoint: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
				mountPoint, stdout, stderr, exit)
			mountPoint = ""
		}
	}

	return &MountInfo{
		MountPoint: mountPoint,
		DeviceID:   devID,
		FsType:     fsType,
	}, err
}

// SetSubtree is used to set quota id for substree dir which is container's root dir.
// For container, it has its own root dir.
// And this dir is a subtree of the host dir which is mapped to a device.
// ext4: chattr -p quotaid +P $DIR
func (quota *PrjQuotaDriver) setQuotaID(dir string, qid uint32, mountInfo *MountInfo) (uint32, error) {
	log.With(nil).Debugf("set subtree, dir: %s, quotaID: %d", dir, qid)

	if isRegular, err := CheckRegularFile(dir); err != nil || !isRegular {
		log.With(nil).Debugf("set quota id skip not regular file: %s", dir)
		return 0, errors.Errorf("file(%s) is not regular file", dir)
	}

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
	log.With(nil).Infof("set quota id, dir: (%s), quota id: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
		dir, strid, stdout, stderr, exit)
	return id, errors.Wrapf(err, "failed to chattr, dir: (%s), quota id: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
		dir, strid, stdout, stderr, exit)
}

// SetDiskQuota uses the following two parameters to set disk quota for a directory.
// * quota size: a byte size of requested quota.
// * quota ID: an ID represent quota attr which is used in the global scope.
func (quota *PrjQuotaDriver) SetDiskQuota(dir string, size string, quotaID uint32) error {
	log.With(nil).Debugf("set disk quota, dir: %s, size: %s, quotaID: %d", dir, size, quotaID)
	mountInfo, err := quota.EnforceQuota(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to enforce quota, dir: (%s)", dir)
	}
	if mountInfo == nil || mountInfo.MountPoint == "" {
		return errors.Errorf("failed to find mountpoint, dir: (%s)", dir)
	}

	// transfer limit from kbyte to byte
	limit, err := bytefmt.ToKilobytes(size)
	if err != nil {
		return errors.Wrapf(err, "failed to change size: (%s) to kilobytes", size)
	}

	if err := checkDevLimit(mountInfo, limit*1024); err != nil {
		return errors.Wrapf(err, "failed to check device limit, dir: (%s), limit: (%d)kb", dir, limit)
	}

	id, err := quota.setQuotaID(dir, quotaID, mountInfo)
	if err != nil {
		return errors.Wrapf(err, "failed to set subtree, dir: (%s), quota id: (%d)", dir, quotaID)
	}
	if id == 0 {
		return errors.Errorf("failed to find quota id to set subtree")
	}

	return quota.setQuota(id, limit, mountInfo)
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

	// /dev/sdb1 /home/pouch ext4 rw,relatime,prjquota,data=ordered 0 0
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

		// check the device turn on the prjquota or not.
		for _, value := range strings.Split(parts[3], ",") {
			if value == "prjquota" {
				enableQuota = true
				break
			}
		}
	}

	log.With(nil).Debugf("check device: (%d), mountpoint: (%s), enableQuota: (%v), fsType: (%s)",
		devID, mountPoint, enableQuota, fsType)

	return mountPoint, enableQuota, fsType
}

// setQuota uses system tool "setquota" to set project quota for binding of limit and mountpoint and quotaID.
// * quotaID: quota ID which means this ID is used in the global scope.
// * blockLimit: block limit number for mountpoint.
// * mountPoint: the mountpoint of the device in the filesystem
// ext4: setquota -P qid $softlimit $hardlimit $softinode $hardinode mountpoint
func (quota *PrjQuotaDriver) setQuota(quotaID uint32, blockLimit uint64, mountInfo *MountInfo) error {
	mountPoint := mountInfo.MountPoint
	log.With(nil).Debugf("set project quota, quotaID: %d, limit: %d, mountpoint: %s", quotaID, blockLimit, mountPoint)

	quotaIDStr := strconv.FormatUint(uint64(quotaID), 10)
	blockLimitStr := strconv.FormatUint(blockLimit, 10)
	// set project quota
	exit, stdout, stderr, err := exec.Run(0, "setquota", "-P", quotaIDStr, "0", blockLimitStr, "0", "0", mountPoint)
	log.With(nil).Infof("set quota size, mountpoint: (%s), quota id: (%d), quota: (%d kbytes), stdout: (%s), stderr: (%s), exit: (%d)",
		mountPoint, quotaID, blockLimit, stdout, stderr, exit)
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
		log.With(nil).Errorf("failed to lsattr, dir: (%s), stdout: (%s), stderr: (%s), exit: (%d), err: (%v)",
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
			log.With(nil).Debugf("get file attr: [%s], quota id: [%d]", dir, qid)
			return uint32(qid)
		}
	}

	log.With(nil).Errorf("failed to get file attr of quota ID for dir %s", dir)
	return 0
}

// SetQuotaIDInFileAttr sets file attributes of quota ID for the input directory.
// The input attributes is quota ID.
func (quota *PrjQuotaDriver) SetQuotaIDInFileAttr(dir string, quotaID uint32) error {
	log.With(nil).Debugf("set file attr, dir: %s, quotaID: %d", dir, quotaID)

	if isRegular, err := CheckRegularFile(dir); err != nil || !isRegular {
		log.With(nil).Debugf("set quota id skip not regular file: %s", dir)
		return errors.Errorf("file(%s) is not regular file", dir)
	}

	strid := strconv.FormatUint(uint64(quotaID), 10)
	exit, stdout, stderr, err := exec.Run(0, "chattr", "-p", strid, "+P", dir)
	return errors.Wrapf(err, "failed to chattr, dir: (%s), quota id: (%d), stdout: (%s), stderr: (%s), exit: (%d)",
		dir, quotaID, stdout, stderr, exit)
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

	log.With(nil).Debugf("get next project quota id: %d", id)
	return id, nil
}

// SetFileAttrRecursive set the file attr by recursively.
func (quota *PrjQuotaDriver) SetFileAttrRecursive(dir string, quotaID uint32) error {
	if isRegular, err := CheckRegularFile(dir); err != nil || !isRegular {
		log.With(nil).Debugf("set quota id skip not regular file: %s", dir)
		return errors.Errorf("file(%s) is not regular file", dir)
	}

	strID := strconv.FormatUint(uint64(quotaID), 10)

	// ext4 use chattr to change project id
	exit, stdout, stderr, err := exec.Run(0, "chattr", "-R", "-p", strID, "+P", dir)
	log.With(nil).Infof("set ext4 project quota id recursively, dir: (%s), quota id: (%s), stdout: (%s), stderr: (%s), exit: (%d)",
		dir, strID, stdout, stderr, exit)
	return errors.Wrapf(err, "failed to set file(%s) quota id(%s) by recursively", dir, strID)
}
