// +build linux

package quota

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/kernel"
	"github.com/alibaba/pouch/pkg/log"
	"github.com/alibaba/pouch/pkg/system"
	"github.com/pkg/errors"
)

const (
	// QuotaMinID represents the minimize quota id.
	// The value is unit32(2^24).
	QuotaMinID = uint32(16777216)

	// procMountFile represent the mounts file in proc virtual file system.
	procMountFile = "/proc/mounts"
)

var (
	// GQuotaDriver represents global quota driver.
	GQuotaDriver = NewQuotaDriver("")
)

// BaseQuota defines the quota operation interface.
// It abstracts the common operation ways a quota driver should implement.
type BaseQuota interface {
	// EnforceQuota is used to enforce disk quota effect on specified directory.
	EnforceQuota(dir string) (*MountInfo, error)

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

	// GetNextQuotaID gets next quota ID in global scope of host.
	GetNextQuotaID() (uint32, error)

	// SetFileAttrRecursive set the file attr by recursively.
	SetFileAttrRecursive(dir string, quotaID uint32) error
}

// NewQuotaDriver returns a quota instance.
func NewQuotaDriver(name string) BaseQuota {
	var quota BaseQuota
	switch name {
	case "grpquota":
		quota = &GrpQuotaDriver{
			quotaIDs: make(map[uint32]struct{}),
		}
	case "prjquota":
		quota = &PrjQuotaDriver{
			quotaIDs: make(map[uint32]struct{}),
		}
	default:
		kernelVersion, err := kernel.GetKernelVersion()
		if err == nil && kernelVersion.Kernel >= 4 {
			quota = &PrjQuotaDriver{
				quotaIDs: make(map[uint32]struct{}),
			}
		} else {
			quota = &GrpQuotaDriver{
				quotaIDs: make(map[uint32]struct{}),
			}
		}
	}

	return quota
}

// SetQuotaDriver is used to set global quota driver.
func SetQuotaDriver(name string) {
	GQuotaDriver = NewQuotaDriver(name)
}

// SetDiskQuota is used to set quota for directory.
func SetDiskQuota(dir string, size string, quotaID uint32) error {
	log.With(nil).Infof("set disk quota, dir(%s), size(%s), quotaID(%d)", dir, size, quotaID)
	if isRegular, err := CheckRegularFile(dir); err != nil || !isRegular {
		log.With(nil).Debugf("set quota skip not regular file: %s", dir)
		return err
	}
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

//GetNextQuotaID returns the next available quota id.
func GetNextQuotaID() (uint32, error) {
	return GQuotaDriver.GetNextQuotaID()
}

// GetQuotaID returns the quota id of directory,
// if no quota id, it will alloc the next available quota id.
func GetQuotaID(dir string) (uint32, error) {
	id := GetQuotaIDInFileAttr(dir)
	if id > 0 {
		return id, nil
	}
	id, err := GetNextQuotaID()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get file(%s) quota id", dir)
	}

	return id, nil
}

// SetRootfsDiskQuota is to set container rootfs dir disk quota.
func SetRootfsDiskQuota(basefs, size string, quotaID uint32, update bool) (uint32, error) {
	overlayMountInfo, err := getOverlayMountInfo(basefs)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get overlay(%s) mount info", basefs)
	}

	for _, dir := range []string{overlayMountInfo.Upper, overlayMountInfo.Work} {
		if quotaID == 0 {
			quotaID, err = GetQuotaID(dir)
			if err != nil {
				return 0, errors.Wrapf(err, "failed to get dir(%s) quota id", dir)
			}
		}

		if err := SetDiskQuota(dir, size, quotaID); err != nil {
			return 0, errors.Wrapf(err, "failed to set dir(%s) disk quota", dir)
		}

		if update {
			go SetFileAttrRecursive(dir, quotaID)
		} else if err := SetFileAttrRecursive(dir, quotaID); err != nil {
			return 0, errors.Wrapf(err, "failed to set dir(%s) quota recursively", dir)
		}
	}

	return quotaID, nil
}

// SetFileAttrRecursive set the file attr by recursively.
func SetFileAttrRecursive(dir string, quotaID uint32) error {
	return GQuotaDriver.SetFileAttrRecursive(dir, quotaID)
}

// CheckRegularFile is used to check the file is regular file or directory.
func CheckRegularFile(file string) (bool, error) {
	fd, err := os.Lstat(file)
	if err != nil {
		log.With(nil).Warnf("failed to check file: %s, err: %v", file, err)
		return false, err
	}

	mode := fd.Mode()
	if mode&(os.ModeSymlink|os.ModeNamedPipe|os.ModeSocket|os.ModeDevice) == 0 {
		return true, nil
	}

	return false, nil
}

// IsSetQuotaID returns whether set quota id
func IsSetQuotaID(id string) bool {
	return id != "" && id != "0"
}

// getOverlayMountInfo gets overlayFS informantion from /proc/mounts.
// upperdir, mergeddir and workdir would be dealt.
func getOverlayMountInfo(basefs string) (*OverlayMount, error) {
	output, err := ioutil.ReadFile(procMountFile)
	if err != nil {
		log.With(nil).Warnf("failed to read file(%s), err(%v)", procMountFile, err)
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
	quotaIDs := make(map[uint32]struct{})

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
	log.With(nil).Infof("Load repquota ids(%d), list(%v)", len(quotaIDs), quotaIDs)
	return quotaIDs, minID, nil
}

// getDevLimit returns the device storage upper limit.
func getDevLimit(info *MountInfo) (uint64, error) {
	mp := info.MountPoint
	devID := info.DeviceID

	newDevID, _ := system.GetDevID(mp)
	if newDevID != devID {
		return 0, errors.Errorf("failed to set device limit, no such device id(%d), checked id(%d)",
			devID, newDevID)
	}

	// get storage upper limit of the device which the dir is on.
	var stfs syscall.Statfs_t
	if err := syscall.Statfs(mp, &stfs); err != nil {
		log.With(nil).Errorf("failed to get path(%s) limit, err(%v)", mp, err)
		return 0, errors.Wrapf(err, "failed to get path(%s) limit", mp)
	}
	limit := stfs.Blocks * uint64(stfs.Bsize)

	log.With(nil).Debugf("get device limit size, mountpoint(%s), limit(%v) B", mp, limit)
	return limit, nil
}

// checkDevLimit checks if the device on which the input dir lies has already been recorded in driver.
func checkDevLimit(mountInfo *MountInfo, size uint64) error {
	mp := mountInfo.MountPoint

	limit, err := getDevLimit(mountInfo)
	if err != nil {
		return errors.Wrapf(err, "failed to get device(%s) limit", mp)
	}

	if limit < size {
		return fmt.Errorf("dir %s quota limit %v must be less than %v", mp, size, limit)
	}

	log.With(nil).Debugf("succeeded in checkDevLimit (dir %s quota limit %v B) with size %v B", mp, limit, size)

	return nil
}

func getDevID(dir string) (uint64, error) {
	// ensure stat syscall don't timeout
	idChan := make(chan uint64)
	errChan := make(chan error)
	timeoutChan := time.After(time.Second * 5)

	go func() {
		id, err := system.GetDevID(dir)
		if err != nil {
			errChan <- err
			return
		}
		idChan <- id
	}()

	select {
	case err := <-errChan:
		return 0, err
	case id := <-idChan:
		return id, nil
	case <-timeoutChan:
		return 0, context.DeadlineExceeded
	}
}
