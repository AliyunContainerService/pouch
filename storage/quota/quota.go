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
	"syscall"

	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/kernel"
	"github.com/alibaba/pouch/pkg/system"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// QuotaMinID represents the minimize quota id.
	// The value is unit32(2^24).
	QuotaMinID = uint32(16777216)

	// procMountFile represent the mounts file in proc virtual file system.
	procMountFile = "/proc/mounts"
)

var (
	hasQuota bool

	// GQuotaDriver represents global quota driver.
	GQuotaDriver = NewQuotaDriver("")

	// devLimits saves all the limit of device.
	// key: device ID
	// value: the storage upper limit size of the device(unit:B)
	devLimits = make(map[uint64]uint64)

	// the lock for devLimits
	lock sync.Mutex
)

// BaseQuota defines the quota operation interface.
// It abstracts the common operation ways a quota driver should implement.
type BaseQuota interface {
	// EnforceQuota is used to enforce disk quota effect on specified directory.
	EnforceQuota(dir string) (string, error)

	// SetSubtree sets quota for container root dir which is a subtree of host's dir mapped on a device.
	SetSubtree(dir string, qid uint32) (uint32, error)

	// SetDiskQuota uses the following two parameters to set disk quota for a directory.
	// * quota size: a byte size of requested quota.
	// * quota ID: an ID represent quota attr which is used in the global scope.
	SetDiskQuota(dir string, size string, quotaID uint32) error

	// CheckMountpoint is used to check mount point.
	// It returns mointpoint, enable quota and filesystem type of the device.
	CheckMountpoint(devID uint64) (string, bool, string)

	// GetQuotaIDInFileAttr gets attributes of the file which is in the inode.
	// The returned result is quota ID.
	GetQuotaIDInFileAttr(dir string) uint32

	// SetQuotaIDInFileAttr sets file attributes of quota ID for the input directory.
	// The input attributes is quota ID.
	SetQuotaIDInFileAttr(dir string, quotaID uint32) error

	// SetQuotaIDInFileAttrNoOutput sets file attributes of quota ID for the input directory without returning error if exists.
	// The input attributes is quota ID.
	SetQuotaIDInFileAttrNoOutput(dir string, quotaID uint32)

	// GetNextQuotaID gets next quota ID in global scope of host.
	GetNextQuotaID() (uint32, error)
}

// NewQuotaDriver returns a quota instance.
func NewQuotaDriver(name string) BaseQuota {
	var quota BaseQuota
	switch name {
	case "grpquota":
		quota = &GrpQuotaDriver{
			quotaIDs:    make(map[uint32]struct{}),
			mountPoints: make(map[uint64]string),
		}
	case "prjquota":
		quota = &PrjQuotaDriver{
			quotaIDs:    make(map[uint32]struct{}),
			mountPoints: make(map[uint64]string),
		}
	default:
		kernelVersion, err := kernel.GetKernelVersion()
		if err == nil && kernelVersion.Kernel >= 4 {
			quota = &PrjQuotaDriver{
				quotaIDs:    make(map[uint32]struct{}),
				mountPoints: make(map[uint64]string),
			}
		} else {
			quota = &GrpQuotaDriver{
				quotaIDs:    make(map[uint32]struct{}),
				mountPoints: make(map[uint64]string),
			}
		}
	}

	return quota
}

// SetQuotaDriver is used to set global quota driver.
func SetQuotaDriver(name string) {
	GQuotaDriver = NewQuotaDriver(name)
}

// StartQuotaDriver is used to start quota driver.
func StartQuotaDriver(dir string) (string, error) {
	return GQuotaDriver.EnforceQuota(dir)
}

// SetSubtree is used to set quota id for directory.
func SetSubtree(dir string, qid uint32) (uint32, error) {
	return GQuotaDriver.SetSubtree(dir, qid)
}

// SetDiskQuota is used to set quota for directory.
func SetDiskQuota(dir string, size string, quotaID uint32) error {
	return GQuotaDriver.SetDiskQuota(dir, size, quotaID)
}

// CheckMountpoint is used to check mount point.
func CheckMountpoint(devID uint64) (string, bool, string) {
	return GQuotaDriver.CheckMountpoint(devID)
}

// GetQuotaIDInFileAttr returns the directory attributes of quota ID.
func GetQuotaIDInFileAttr(dir string) uint32 {
	return GQuotaDriver.GetQuotaIDInFileAttr(dir)
}

// SetQuotaIDInFileAttr is used to set file attributes of quota ID.
func SetQuotaIDInFileAttr(dir string, id uint32) error {
	return GQuotaDriver.SetQuotaIDInFileAttr(dir, id)
}

// SetQuotaIDInFileAttrNoOutput is used to set file attribute of quota ID without error.
func SetQuotaIDInFileAttrNoOutput(dir string, quotaID uint32) {
	GQuotaDriver.SetQuotaIDInFileAttrNoOutput(dir, quotaID)
}

//GetNextQuotaID returns the next available quota id.
func GetNextQuotaID() (uint32, error) {
	return GQuotaDriver.GetNextQuotaID()
}

//GetDefaultQuota returns the default quota size.
func GetDefaultQuota(quotas map[string]string) string {
	if quotas == nil {
		return ""
	}

	// "/" means the disk quota only takes effect on rootfs + 0 * volume
	quota, ok := quotas["/"]
	if ok && quota != "" {
		return quota
	}

	// ".*" means the disk quota only takes effect on rootfs + n * volume
	quota, ok = quotas[".*"]
	if ok && quota != "" {
		return quota
	}

	return ""
}

// SetRootfsDiskQuota is to set container rootfs dir disk quota.
func SetRootfsDiskQuota(basefs, size string, quotaID uint32) (uint32, error) {
	overlayMountInfo, err := getOverlayMountInfo(basefs)
	if err != nil {
		return 0, fmt.Errorf("failed to get overlay mount info: %v", err)
	}

	for _, dir := range []string{overlayMountInfo.Upper, overlayMountInfo.Work} {
		_, err = StartQuotaDriver(dir)
		if err != nil {
			return 0, fmt.Errorf("failed to start quota driver: %v", err)
		}

		quotaID, err = SetSubtree(dir, quotaID)
		if err != nil {
			return 0, fmt.Errorf("failed to set subtree: %v", err)
		}

		if err := SetDiskQuota(dir, size, quotaID); err != nil {
			return 0, fmt.Errorf("failed to set disk quota: %v", err)
		}
	}

	return quotaID, nil
}

// SetQuotaForDir sets file attribute
func SetQuotaForDir(src string, quotaID uint32) error {
	filepath.Walk(src, func(path string, fd os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("setQuota walk dir %s get error %v", path, err)
		}

		SetQuotaIDInFileAttrNoOutput(path, quotaID)
		return nil
	})

	return nil
}

// getOverlayMountInfo gets overlayFS informantion from /proc/mounts.
// upperdir, mergeddir and workdir would be dealt.
func getOverlayMountInfo(basefs string) (*OverlayMount, error) {
	output, err := ioutil.ReadFile(procMountFile)
	if err != nil {
		logrus.Warnf("failed to ReadFile %s: %v", procMountFile, err)
		return nil, err
	}

	var lowerDir, upperDir, workDir string
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(string(line), " ")
		if len(parts) != 6 {
			continue
		}
		if parts[1] != basefs || parts[2] != "overlay" {
			continue
		}
		// the expected format is like following:
		// overlay /var/lib/pouch/containerd/state/io.containerd.runtime.v1.linux/default/8d849ee68c8698531a2575f890be027dbd4dcb64f39cce37d7d22a703cbb362b/rootfs overlay rw,relatime,lowerdir=/var/lib/pouch/containerd/root/io.containerd.snapshotter.v1.overlayfs/snapshots/1/fs,upperdir=/var/lib/pouch/containerd/root/io.containerd.snapshotter.v1.overlayfs/snapshots/274/fs,workdir=/var/lib/pouch/containerd/root/io.containerd.snapshotter.v1.overlayfs/snapshots/274/work 0 0
		// In part[3], it stored lowerdir, upperdir and workdir.
		mountParams := strings.Split(parts[3], ",")
		for _, p := range mountParams {
			switch {
			case strings.Contains(p, "lowerdir"):
				if s := strings.Split(p, "="); len(s) == 2 {
					lowerDir = s[1]
				}
			case strings.Contains(p, "upperdir"):
				if s := strings.Split(p, "="); len(s) == 2 {
					upperDir = s[1]
				}
			case strings.Contains(p, "workdir"):
				if s := strings.Split(p, "="); len(s) == 2 {
					workDir = s[1]
					break
				}
			}
		}
	}

	if lowerDir == "" || upperDir == "" || workDir == "" {
		return nil, fmt.Errorf("failed to get OverlayFs Mount Info: lowerdir, upperdir, workdir must be non-empty")
	}

	return &OverlayMount{
		Lower: lowerDir,
		Upper: upperDir,
		Work:  workDir,
	}, nil
}

// loadQuotaIDs loads quota IDs for quota driver from reqquota execution result.
// This function utils `repquota` which summarizes quotas for a filesystem.
// see http://man7.org/linux/man-pages/man8/repquota.8.html
//
// $ repquota -Pan
// Project         used    soft    hard  grace    used  soft  hard  grace
// ----------------------------------------------------------------------
// #0        --     220       0       0             25     0     0
// #123      --       4       0 88589934592          1     0     0
// #8888     --       8       0       0              2     0     0
//
// Or
//
// $ repquota -gan
// Group           used    soft    hard  grace    used  soft  hard  grace
// ----------------------------------------------------------------------
// #0        --  494472       0       0            938     0     0
// #54       --       8       0       0              2     0     0
// #4        --      16       0       0              4     0     0
// #22       --      28       0       0              4     0     0
// #16777220 +- 2048576       0 2048575              9     0     0
// #500      --   47504       0       0            101     0     0
// #16777221 -- 3048576       0 3048576              8     0     0
func loadQuotaIDs(repquotaOpt string) (map[uint32]struct{}, uint32, error) {
	quotaIDs := make(map[uint32]struct{}, 0)

	minID := QuotaMinID
	exit, output, stderr, err := exec.Run(0, "repquota", repquotaOpt)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to execute [repquota %s], stdout: (%s), stderr: (%s), exit: (%d)",
			repquotaOpt, output, stderr, exit)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) == 0 || line[0] != '#' {
			continue
		}
		// find all lines with prefix '#'
		parts := strings.Split(line, " ")
		// part[0] is "#123456"
		if len(parts[0]) <= 1 {
			continue
		}

		id, err := strconv.Atoi(parts[0][1:])
		quotaID := uint32(id)
		if err == nil && quotaID > QuotaMinID {
			quotaIDs[quotaID] = struct{}{}
			if quotaID > minID {
				minID = quotaID
			}
		}
	}
	logrus.Infof("Load repquota ids: %d, list: %v", len(quotaIDs), quotaIDs)
	return quotaIDs, minID, nil
}

func getMountpoint(dir string) (string, error) {
	var (
		mountPoint string
	)

	output, err := ioutil.ReadFile(procMountFile)
	if err != nil {
		logrus.Warnf("failed to read file: (%s), err: (%v)", procMountFile, err)
		return "", errors.Wrapf(err, "failed to read file: (%s)", procMountFile)
	}

	devID, err := system.GetDevID(dir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get device id for dir: (%s)", dir)
	}
	logrus.Debugf("get directory(%s) device id: (%d)", dir, devID)

	// /dev/sdb1 /home/pouch ext4 rw,relatime,prjquota,data=ordered 0 0
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) != 6 {
			continue
		}

		// only check xfs/ext3/ext4 file system
		if parts[2] != "xfs" && parts[2] != "ext3" && parts[2] != "ext4" {
			continue
		}

		newDevID, err := system.GetDevID(parts[1])
		if err != nil {
			continue
		}

		// /dev/sdb1 /home/pouch ext4 rw,relatime,prjquota,data=ordered 0 0
		// /dev/sdb1 /home/pouch/overlay ext4 rw,relatime,prjquota,data=ordered 0 0
		// we will choose the shortest match string, /home/pouch
		if devID == newDevID && strings.HasPrefix(dir, parts[1]) &&
			(mountPoint == "" || len(parts[1]) < len(mountPoint)) {
			mountPoint = parts[1]
			devID = newDevID
		}
	}

	if mountPoint == "" {
		return "", errors.Errorf("failed to get mount point of directory: (%s)", dir)
	}

	logrus.Debugf("get the directory: (%s) mountpoint: (%s)", dir, mountPoint)

	return mountPoint, nil
}

// setDevLimit sets device storage upper limit in quota driver according to inpur dir.
func setDevLimit(dir string, devID uint64) (uint64, error) {
	lock.Lock()
	limit, exist := devLimits[devID]
	lock.Unlock()
	if exist {
		return limit, nil
	}

	mp, err := getMountpoint(dir)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to set device limit, dir: (%s), devID: (%d)", dir, devID)
	}

	newDevID, _ := system.GetDevID(mp)
	if newDevID != devID {
		return 0, errors.Errorf("failed to set device limit, no such device id: (%d), checked id: (%d)",
			devID, newDevID)
	}

	// get storage upper limit of the device which the dir is on.
	var stfs syscall.Statfs_t
	if err := syscall.Statfs(mp, &stfs); err != nil {
		logrus.Errorf("failed to get path: (%s) limit, err: (%v)", mp, err)
		return 0, errors.Wrapf(err, "failed to get path: (%s) limit", mp)
	}
	limit = stfs.Blocks * uint64(stfs.Bsize)

	lock.Lock()
	devLimits[devID] = limit
	lock.Unlock()

	logrus.Debugf("SetDevLimit: dir: (%s), mountpoint: (%s), limit: (%v) B", dir, mp, limit)
	return limit, nil
}

// checkDevLimit checks if the device on which the input dir lies has already been recorded in driver.
func checkDevLimit(dir string, size uint64) error {
	devID, err := system.GetDevID(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to get device id, dir: (%s)", dir)
	}

	lock.Lock()
	limit, exist := devLimits[devID]
	lock.Unlock()
	if !exist {
		// if has not recorded, just add (dir, device, limit) to driver.
		if limit, err = setDevLimit(dir, devID); err != nil {
			return errors.Wrapf(err, "failed to set device limit, dir: (%s), devID: (%d)", dir, devID)
		}
	}

	if limit < size {
		return fmt.Errorf("dir %s quota limit %v must be less than %v", dir, size, limit)
	}

	logrus.Debugf("succeeded in checkDevLimit (dir %s quota limit %v B) with size %v B", dir, limit, size)

	return nil
}
