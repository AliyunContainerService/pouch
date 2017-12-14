package environment

import (
	"runtime"

	"github.com/alibaba/pouch/pkg/utils"
)

var (
	// PouchBinary is default binary
	PouchBinary = "/usr/local/bin/pouch"

	// PouchdAddress is default pouchd address
	PouchdAddress = "unix:///var/run/pouchd.sock"

	// TLSConfig is default tls config
	TLSConfig = utils.TLSConfig{}
)

// IsLinux checks if the OS of test environment is Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
