package quota

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/alibaba/pouch/pkg/bytefmt"
	"github.com/alibaba/pouch/pkg/exec"

	"github.com/sirupsen/logrus"
)

// GrpQuota represents group quota.
type GrpQuota struct {
	lock sync.Mutex
	// quotaIDs saves all of quota ids.
	quotaIDs map[uint32]uint32
	// mountPoints saves all the mount point of volume.
	mountPoints map[uint64]string
	// quotaLastID is used to mark last used quota id.
	quotaLastID uint32
}

// StartQuotaDriver is used to start quota driver.
func (quota *GrpQuota) StartQuotaDriver(dir string) (string, error) {
	logrus.Debugf("start group quota driver: %s", dir)
	if !UseQuota {
		return "", nil
	}

	devID, err := GetDevID(dir)
	if err != nil {
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
		_, _, _, err := exec.Run(0, "mount", "-o", "remount,grpquota", mountPoint)
		if err != nil {
			return "", err
		}
	}

	vfsVersion, quotaFilename, err := quota.getVFSVersionAndQuotaFile(devID)
	if err != nil {
		return "", err
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

		if writeErr := ioutil.WriteFile(filename, header, 644); writeErr != nil {
			logrus.Errorf("write file error. %s, %s, %s", filename, vfsVersion, writeErr)
			return mountPoint, writeErr
		}
		if _, _, _, err := exec.Run(0, "setquota", "-g", "-t", "43200", "43200", mountPoint); err != nil {
			os.Remove(filename)
			return mountPoint, err
		}
		if err := quota.setUserQuota(0, 0, mountPoint); err != nil {
			os.Remove(filename)
			return mountPoint, err
		}
	}

	// check group quota status, on or not, pay attention, the right exit code of command 'quotaon' is '1'.
	exit, stdout, stderr, err := exec.Run(0, "quotaon", "-pg", mountPoint)
	if err != nil && exit != 1 {
		logrus.Errorf("quotaon failed, exit: %d, stdout: %s, stderr: %s, err: %v", exit, stdout, stderr, err)
		return "", fmt.Errorf("stderr: %s, err: %v", stderr, err)
	}
	if strings.Contains(stdout, " is on") {
		quota.mountPoints[devID] = mountPoint
		return mountPoint, nil
	}
	if _, _, _, err = exec.Run(0, "quotaon", mountPoint); err != nil {
		mountPoint = ""
	}

	quota.mountPoints[devID] = mountPoint
	return mountPoint, err
}

// SetSubtree is used to set quota id for directory,
// setfattr -n system.subtree -v $QUOTAID
func (quota *GrpQuota) SetSubtree(dir string, qid uint32) (uint32, error) {
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
	_, _, _, err = exec.Run(0, "setfattr", "-n", "system.subtree", "-v", strid, dir)
	return id, err
}

// SetDiskQuota is used to set quota for directory.
func (quota *GrpQuota) SetDiskQuota(dir string, size string, quotaID uint32) error {
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
		return err
	}

	return quota.setUserQuota(id, limit, mountPoint)
}

// CheckMountpoint is used to check mount point.
// cat /proc/mounts as follows:
// /dev/sda3 / ext4 rw,relatime,data=ordered 0 0
// /dev/sda2 /boot/grub2 ext4 rw,relatime,stripe=4,data=ordered 0 0
// /dev/sda5 /home ext4 rw,relatime,data=ordered 0 0
// /dev/sdb1 /home/pouch ext4 rw,relatime,grpquota,data=ordered 0 0
// tmpfs /run tmpfs rw,nosuid,nodev,mode=755 0 0
// tmpfs /sys/fs/cgroup tmpfs ro,nosuid,nodev,noexec,mode=755 0 0
// cgroup /sys/fs/cgroup/cpuset,cpu,cpuacct cgroup rw,nosuid,nodev,noexec,relatime,cpuacct,cpu,cpuset 0 0
// cgroup /sys/fs/cgroup/devices cgroup rw,nosuid,nodev,noexec,relatime,devices 0 0
// cgroup /sys/fs/cgroup/memory cgroup rw,nosuid,nodev,noexec,relatime,memory 0 0
// cgroup /sys/fs/cgroup/blkio cgroup rw,nosuid,nodev,noexec,relatime,blkio 0 0
func (quota *GrpQuota) CheckMountpoint(devID uint64) (string, bool, string) {
	logrus.Debugf("check mountpoint, devID: %d", devID)
	output, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		logrus.Warnf("ReadFile: %v", err)
		return "", false, ""
	}

	var mountPoint, fsType string
	hasQuota := false
	// /dev/sdb1 /home/pouch ext4 rw,relatime,grpquota,data=ordered 0 0
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
				if opt == "grpquota" {
					hasQuota = true
				}
			}
			break
		}
	}

	return mountPoint, hasQuota, fsType
}

func (quota *GrpQuota) setUserQuota(quotaID uint32, diskQuota uint64, mountPoint string) error {
	logrus.Debugf("set user quota, quotaID: %d, limit: %d, mountpoint: %s", quotaID, diskQuota, mountPoint)

	uid := strconv.FormatUint(uint64(quotaID), 10)
	limit := strconv.FormatUint(diskQuota, 10)

	_, _, _, err := exec.Run(0, "setquota", "-g", uid, "0", limit, "0", "0", mountPoint)
	return err
}

// GetFileAttr returns the directory attributes
// getfattr -n system.subtree --only-values --absolute-names /
func (quota *GrpQuota) GetFileAttr(dir string) uint32 {
	logrus.Debugf("get file attr, dir: %s", dir)

	v := 0
	_, out, _, err := exec.Run(0, "getfattr", "-n", "system.subtree", "--only-values", "--absolute-names", dir)
	if err == nil {
		v, _ = strconv.Atoi(out)
	}
	return uint32(v)
}

// SetFileAttr is used to set file attributes.
func (quota *GrpQuota) SetFileAttr(dir string, id uint32) error {
	logrus.Debugf("set file attr, dir: %s, quotaID: %d", dir, id)

	strid := strconv.FormatUint(uint64(id), 10)
	_, _, _, err := exec.Run(0, "setfattr", "-n", "system.subtree", "-v", strid, dir)
	return err
}

// SetFileAttrNoOutput is used to set file attributes without error.
func (quota *GrpQuota) SetFileAttrNoOutput(dir string, id uint32) {
	strid := strconv.FormatUint(uint64(id), 10)
	exec.Run(0, "setfattr", "-n", "system.subtree", "-v", strid, dir)
}

// load
// repquota -gan
// Group           used    soft    hard  grace    used  soft  hard  grace
// ----------------------------------------------------------------------
// #0        --  494472       0       0            938     0     0
// #54       --       8       0       0              2     0     0
// #4        --      16       0       0              4     0     0
// #22       --      28       0       0              4     0     0
// #16777220 +- 2048576       0 2048575              9     0     0
// #500      --   47504       0       0            101     0     0
// #16777221 -- 3048576       0 3048576              8     0     0
func (quota *GrpQuota) loadQuotaIDs() (uint32, error) {
	minID := QuotaMinID
	_, stdout, _, err := exec.Run(0, "repquota", "-gan")
	if err != nil {
		return minID, err
	}

	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		if len(line) == 0 || line[0] != '#' {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
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
func (quota *GrpQuota) GetNextQuatoID() (uint32, error) {
	quota.lock.Lock()
	defer quota.lock.Unlock()

	if quota.quotaLastID == 0 {
		var err error
		quota.quotaLastID, err = quota.loadQuotaIDs()
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

func (quota *GrpQuota) getVFSVersionAndQuotaFile(devID uint64) (string, string, error) {
	output, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		logrus.Warnf("ReadFile: %v", err)
		return "", "", err
	}

	vfsVersion := "vfsv0"
	quotaFilename := "aquota.group"
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) != 6 {
			continue
		}

		devID2, _ := GetDevID(parts[1])
		if devID == devID2 {
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
			break
		}
	}

	return vfsVersion, quotaFilename, nil
}
