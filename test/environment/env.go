package environment

import (
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
