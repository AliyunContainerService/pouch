package kernel

import (
	"fmt"

	"github.com/alibaba/pouch/pkg/exec"
	"github.com/pkg/errors"
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

	_, stdout, _, err := exec.Run(0, "uname", "-r")
	if err != nil {
		return nil, errors.Wrap(err, "failed to run command uname -r")
	}

	parsed, _ := fmt.Sscanf(stdout, "%d.%d.%d-%s", &kernel, &major, &minor, &flavor)
	if parsed < 3 {
		return nil, fmt.Errorf("Can't parse kernel version, release: %s" + stdout)
	}

	return &VersionInfo{
		Kernel: kernel,
		Major:  major,
		Minor:  minor,
		Flavor: flavor,
	}, nil
}
