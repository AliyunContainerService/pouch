package environment

import (
	"os"
	"runtime"

	"github.com/alibaba/pouch/pkg/utils"
	"github.com/gotestyourself/gotestyourself/icmd"
)

var (
	// PouchBinary is default binary
	PouchBinary = "/usr/local/bin/pouch"

	// PouchdAddress is default pouchd address
	PouchdAddress = "unix:///var/run/pouchd.sock"

	// PouchdUnixDomainSock is the default unix domain socket file used by pouchd.
	PouchdUnixDomainSock = "/var/run/pouchd.sock"

	// TLSConfig is default tls config
	TLSConfig = utils.TLSConfig{}
)

// IsLinux checks if the OS of test environment is Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsAliKernel checks if the kernel of test environment is AliKernel.
func IsAliKernel() bool {
	cmd := "uname -r | grep -i alios"
	if icmd.RunCommand("bash", "-c", cmd).ExitCode == 0 {
		return true
	}
	return false
}

// IsDumbInitExist checks if the dumb-init binary exists on host.
func IsDumbInitExist() bool {
	if _, err := os.Stat("/usr/bin/dumb-init"); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// IsRuncVersionSupportRichContianer checks if the version of runc supports rich container.
func IsRuncVersionSupportRichContianer() bool {
	cmd := "runc -v|grep 1.0.0-rc4-1"
	if icmd.RunCommand("bash", "-c", cmd).ExitCode == 0 {
		return true
	}
	return false
}

// IsHubConnected checks if hub address can be connected.
func IsHubConnected() bool {
	// TODO: found a proper way to test if hub address can be connected.
	return true
}

// IsDiskQuota checks if it can use disk quota for container.
func IsDiskQuota() bool {
	if icmd.RunCommand("which", "quotaon").ExitCode == 0 {
		return true
	}
	return false
}
