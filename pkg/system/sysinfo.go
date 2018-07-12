// +build linux

package system

import (
	"syscall"
)

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
