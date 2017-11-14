package local

import (
	"syscall"

	"github.com/sirupsen/logrus"
)

const (
	// QuotaMinID represents the minimize quota id.
	QuotaMinID = uint32(16777216)
)

var (
	// UseQuota represents use quota or not.
	UseQuota = true
	gquota   = NewQuota("")
)

// BaseQuota defines the quota operation interface.
type BaseQuota interface {
	StartQuotaDriver(dir string) (string, error)
	SetSubtree(dir string, qid uint32) (uint32, error)
	SetDiskQuota(dir string, size string, quotaID int) error
	CheckMountpoint(devID uint64) (string, bool, string)
	GetFileAttr(dir string) uint32
	SetFileAttr(dir string, id uint32) error
	SetFileAttrNoOutput(dir string, id uint32)
	GetNextQuatoID() (uint32, error)
}

// NewQuota returns a quota instance.
func NewQuota(name string) BaseQuota {
	var quota BaseQuota
	switch name {
	case "grp":
		quota = &GrpQuota{
			quotaIDs:    make(map[uint32]uint32),
			mountPoints: make(map[uint64]string),
		}
	case "prj":
		quota = &PrjQuota{
			quotaIDs:    make(map[uint32]uint32),
			mountPoints: make(map[uint64]string),
			devLimits:   make(map[uint64]uint64),
		}
	default:
		quota = &GrpQuota{
			quotaIDs:    make(map[uint32]uint32),
			mountPoints: make(map[uint64]string),
		}
	}

	return quota
}

// GetDevID returns device id.
func GetDevID(dir string) (uint64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(dir, &st); err != nil {
		logrus.Warnf("getDirDev: %s, %v", dir, err)
		return 0, err
	}
	return st.Dev, nil
}

// StartQuotaDriver is used to start quota driver.
func StartQuotaDriver(dir string) (string, error) {
	return gquota.StartQuotaDriver(dir)
}

// SetSubtree is used to set quota id for directory.
func SetSubtree(dir string, qid uint32) (uint32, error) {
	return gquota.SetSubtree(dir, qid)
}

// SetDiskQuota is used to set quota for directory.
func SetDiskQuota(dir string, size string, quotaID int) error {
	return gquota.SetDiskQuota(dir, size, quotaID)
}

// CheckMountpoint is used to check mount point.
func CheckMountpoint(devID uint64) (string, bool, string) {
	return gquota.CheckMountpoint(devID)
}

// GetFileAttr returns the directory attributes.
func GetFileAttr(dir string) uint32 {
	return gquota.GetFileAttr(dir)
}

// SetFileAttr is used to set file attributes.
func SetFileAttr(dir string, id uint32) error {
	return gquota.SetFileAttr(dir, id)
}

// SetFileAttrNoOutput is used to set file attributes without error.
func SetFileAttrNoOutput(dir string, id uint32) {
	gquota.SetFileAttrNoOutput(dir, id)
}

//GetNextQuatoID returns the next available quota id.
func GetNextQuatoID() (uint32, error) {
	return gquota.GetNextQuatoID()
}
