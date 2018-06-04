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

	"github.com/sirupsen/logrus"
)

// PrjQuota represents project quota.
type PrjQuota struct {
	lock sync.Mutex
	// quotaIDs saves all of quota ids.
	quotaIDs map[uint32]uint32
	// mountPoints saves all the mount point of volume.
	mountPoints map[uint64]string
	// devLimits saves all the limit of device.
	devLimits map[uint64]uint64
	// quotaLastID is used to mark last used quota id.
	quotaLastID uint32
}

// StartQuotaDriver is used to start quota driver.
func (quota *PrjQuota) StartQuotaDriver(dir string) (string, error) {
	logrus.Debugf("start project quota driver: %s", dir)
	if !UseQuota {
		return "", nil
	}

	devID, err := GetDevID(dir)
	if err != nil {
		return "", err
	}

	if _, err = quota.setDevLimit(dir, devID); err != nil {
		return "", err
	}

	quota.lock.Lock()
	defer quota.lock.Unlock()

	if mp, ok := quota.mountPoints[devID]; ok {
		return mp, nil
	}

	mountPoint, hasQuota, _ := quota.CheckMountpoint(devID)
	if len(mountPoint) == 0 {
		return mountPoint, fmt.Errorf("mountPoint not found: %s", dir)
	}
	if !hasQuota {
		exec.Run(0, "mount", "-o", "remount,prjquota", mountPoint)
	}

	// on
	_, _, stderr, err := exec.Run(0, "quotaon", "-P", mountPoint)
	if err != nil {
		if strings.Contains(stderr, " File exists") {
			err = nil
		} else {
			mountPoint = ""
		}
	}

	quota.mountPoints[devID] = mountPoint
	return mountPoint, err
}

// SetSubtree is used to set quota id for directory,
//chattr -p qid +P $QUOTAID
func (quota *PrjQuota) SetSubtree(dir string, qid uint32) (uint32, error) {
	logrus.Debugf("set subtree, dir: %s, quotaID: %d", dir, qid)
	if !UseQuota {
		return 0, nil
	}

	id := qid
	var err error
	if id == 0 {
		id = quota.GetFileAttr(dir)
		if id > 0 {
			return id, nil
		}
		id, err = quota.GetNextQuatoID()
	}

	if err != nil {
		return 0, err
	}
	strid := strconv.FormatUint(uint64(id), 10)
	_, _, _, err = exec.Run(0, "chattr", "-p", strid, "+P", dir)
	return id, err
}

// SetDiskQuota is used to set quota for directory.
func (quota *PrjQuota) SetDiskQuota(dir string, size string, quotaID uint32) error {
	logrus.Debugf("set disk quota, dir: %s, size: %s, quotaID: %d", dir, size, quotaID)
	if !UseQuota {
		return nil
	}

	mountPoint, err := quota.StartQuotaDriver(dir)
	if err != nil {
		return err
	}
	if len(mountPoint) == 0 {
		return fmt.Errorf("mountpoint not found: %s", dir)
	}

	id, err := quota.SetSubtree(dir, quotaID)
	if err != nil || id == 0 {
		return fmt.Errorf("subtree not found: %s %v", dir, err)
	}

	limit, err := bytefmt.ToKilobytes(size)
	if err != nil {
		return fmt.Errorf("invalid size: %s %v", size, err)
	}

	// transfer limit from kbyte to byte
	if err := quota.checkDevLimit(dir, limit*1024); err != nil {
		return err
	}

	return quota.setUserQuota(id, limit, mountPoint)
}

// CheckMountpoint is used to check mount point.
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
func (quota *PrjQuota) CheckMountpoint(devID uint64) (string, bool, string) {
	logrus.Debugf("check mountpoint, devID: %d", devID)
	output, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		logrus.Warnf("ReadFile: %v", err)
		return "", false, ""
	}

	var mountPoint, fsType string
	hasQuota := false
	// /dev/sdb1 /home/pouch ext4 rw,relatime,prjquota,data=ordered 0 0
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) != 6 {
			continue
		}
		devID2, _ := GetDevID(parts[1])

		if devID == devID2 {
			mountPoint = parts[1]
			fsType = parts[2]
			for _, opt := range strings.Split(parts[3], ",") {
				if opt == "prjquota" {
					hasQuota = true
				}
			}
			break
		}
	}

	return mountPoint, hasQuota, fsType
}

func (quota *PrjQuota) setUserQuota(quotaID uint32, diskQuota uint64, mountPoint string) error {
	logrus.Debugf("set user quota, quotaID: %d, limit: %d, mountpoint: %s", quotaID, diskQuota, mountPoint)

	uid := strconv.FormatUint(uint64(quotaID), 10)
	limit := strconv.FormatUint(diskQuota, 10)
	_, _, _, err := exec.Run(0, "setquota", "-P", uid, "0", limit, "0", "0", mountPoint)
	return err
}

// GetFileAttr returns the directory attributes
// lsattr -p $dir
func (quota *PrjQuota) GetFileAttr(dir string) uint32 {
	parent := path.Dir(dir)
	qid := 0

	_, out, _, err := exec.Run(0, "lsattr", "-p", parent)
	if err != nil {
		return 0
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) > 2 && parts[2] == dir {
			qid, _ = strconv.Atoi(parts[0])
			break
		}
	}

	logrus.Debugf("get file attr: [%s], quota id: [%d]", dir, qid)

	return uint32(qid)
}

// SetFileAttr is used to set file attributes.
func (quota *PrjQuota) SetFileAttr(dir string, id uint32) error {
	logrus.Debugf("set file attr, dir: %s, quotaID: %d", dir, id)

	strid := strconv.FormatUint(uint64(id), 10)
	_, _, _, err := exec.Run(0, "chattr", "-p", strid, "+P", dir)
	return err
}

// SetFileAttrNoOutput is used to set file attributes without error.
func (quota *PrjQuota) SetFileAttrNoOutput(dir string, id uint32) {
	strid := strconv.FormatUint(uint64(id), 10)
	exec.Run(0, "chattr", "-p", strid, "+P", dir)
}

// load
// repquota -Pan
// Project         used    soft    hard  grace    used  soft  hard  grace
// ----------------------------------------------------------------------
// #0        --     220       0       0             25     0     0
// #123      --       4       0 88589934592          1     0     0
// #8888     --       8       0       0              2     0     0
func (quota *PrjQuota) loadQuotaIds() (uint32, error) {
	minID := QuotaMinID
	_, output, _, err := exec.Run(0, "repquota", "-Pan")
	if err != nil {
		return minID, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) == 0 || line[0] != '#' {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) == 0 {
			continue
		}
		id, err := strconv.Atoi(parts[0][1:])
		uid := uint32(id)
		if err == nil && uid > QuotaMinID {
			quota.quotaIDs[uid] = 1
			if uid > minID {
				minID = uid
			}
		}
	}
	logrus.Infof("Load repquota ids: %d, list: %v", len(quota.quotaIDs), quota.quotaIDs)
	return minID, nil
}

// GetNextQuatoID returns the next available quota id.
func (quota *PrjQuota) GetNextQuatoID() (uint32, error) {
	quota.lock.Lock()
	defer quota.lock.Unlock()

	if quota.quotaLastID == 0 {
		var err error
		quota.quotaLastID, err = quota.loadQuotaIds()
		if err != nil {
			return 0, err
		}
	}
	id := quota.quotaLastID
	for {
		if id < QuotaMinID {
			id = QuotaMinID
		}
		id++
		if _, ok := quota.quotaIDs[id]; !ok {
			break
		}
	}
	quota.quotaIDs[id] = 1
	quota.quotaLastID = id

	logrus.Debugf("get next project quota id: %d", id)
	return id, nil
}

func (quota *PrjQuota) setDevLimit(dir string, devID uint64) (uint64, error) {
	if limit, exist := quota.devLimits[devID]; exist {
		return limit, nil
	}

	var stfs syscall.Statfs_t
	if err := syscall.Statfs(dir, &stfs); err != nil {
		logrus.Errorf("fail to get path %s limit: %v", dir, err)
		return 0, err
	}

	limit := stfs.Blocks * uint64(stfs.Bsize)

	quota.lock.Lock()
	quota.devLimits[devID] = limit
	quota.lock.Unlock()

	logrus.Debugf("SetDevLimit: dir %s limit is %v B", dir, limit)
	return limit, nil
}

func (quota *PrjQuota) checkDevLimit(dir string, size uint64) error {
	devID, err := GetDevID(dir)
	if err != nil {
		return err
	}

	limit, exist := quota.devLimits[devID]
	if !exist {
		if limit, err = quota.setDevLimit(dir, devID); err != nil {
			return err
		}
	}

	if limit < size {
		return fmt.Errorf("dir %s quota limit should < %v, use exceed limit %v", dir, limit, size)
	}

	logrus.Debugf("checkDevLimit dir %s quota limit %v B, size %v B", dir, limit, size)

	return nil
}
