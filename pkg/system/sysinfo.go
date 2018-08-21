// +build linux

package system

import (
	"io/ioutil"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// SysInfo defines system info on current machine
type SysInfo struct {
	AppArmor bool
	Seccomp  bool

	*CgroupInfo
}

// NewSysInfo creates a system info about current machine.
func NewSysInfo() *SysInfo {
	sysInfo := &SysInfo{CgroupInfo: NewCgroupInfo()}

	// Check if AppArmor is supported.
	// isAppArmorEnabled returns true if apparmor is enabled for the host.
	// This function is forked from
	// https://github.com/opencontainers/runc/blob/1a81e9ab1f138c091fe5c86d0883f87716088527/libcontainer/apparmor/apparmor.go
	// to avoid the libapparmor dependency.
	if _, err := os.Stat("/sys/kernel/security/apparmor"); err == nil && os.Getenv("container") == "" {
		if _, err = os.Stat("/sbin/apparmor_parser"); err == nil {
			buf, err := ioutil.ReadFile("/sys/module/apparmor/parameters/enabled")
			if err == nil && len(buf) > 1 && buf[0] == 'Y' {
				sysInfo.AppArmor = true
			}
		}
	}

	// Check if Seccomp is supported, via CONFIG_SECCOMP.
	if err := unix.Prctl(unix.PR_GET_SECCOMP, 0, 0, 0, 0); err != unix.EINVAL {
		// Make sure the kernel has CONFIG_SECCOMP_FILTER.
		if err := unix.Prctl(unix.PR_SET_SECCOMP, unix.SECCOMP_MODE_FILTER, 0, 0, 0); err != unix.EINVAL {
			sysInfo.Seccomp = true
		}
	}

	return sysInfo
}

// getSysInfo gets sysinfo.
func getSysInfo() (*syscall.Sysinfo_t, error) {
	si := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(si)
	if err != nil {
		return nil, err
	}
	return si, nil
}

// GetTotalMem gets total ram of host.
func GetTotalMem() (uint64, error) {
	si, err := getSysInfo()
	if err != nil {
		return 0, err
	}
	return si.Totalram, nil
}
