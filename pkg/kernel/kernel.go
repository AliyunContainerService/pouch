package kernel

import (
	"bytes"
	"fmt"

	"golang.org/x/sys/unix"
)

// VersionInfo holds information about the kernel.
type VersionInfo struct {
	Kernel int    // Version of the kernel (e.g. 4.1.2-generic -> 4)
	Major  int    // Major part of the kernel version (e.g. 4.1.2-generic -> 1)
	Minor  int    // Minor part of the kernel version (e.g. 4.1.2-generic -> 2)
	Flavor string // Flavor of the kernel version (e.g. 4.1.2-generic -> generic)
}

// String returns the kernel version's string format.
func (k *VersionInfo) String() string {
	return fmt.Sprintf("%d.%d.%d-%s", k.Kernel, k.Major, k.Minor, k.Flavor)
}

// GetKernelVersion returns the kernel version info.
func GetKernelVersion() (*VersionInfo, error) {
	var (
		kernel, major, minor int
		flavor               string
	)

	buf := unix.Utsname{}
	err := unix.Uname(&buf)
	if err != nil {
		return nil, err
	}
	// Remove \x00 from Release
	release := string(buf.Release[:bytes.IndexByte(buf.Release[:], 0)])
	parsed, _ := fmt.Sscanf(release, "%d.%d.%d-%s", &kernel, &major, &minor, &flavor)
	if parsed < 3 {
		return nil, fmt.Errorf("Can't parse kernel version, release: %s" + release)
	}

	return &VersionInfo{
		Kernel: kernel,
		Major:  major,
		Minor:  minor,
		Flavor: flavor,
	}, nil
}
