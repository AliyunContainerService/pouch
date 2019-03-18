package daemonv2

import (
	"os"

	daemon "github.com/alibaba/pouch/daemon/config"
)

// For pouch daemon test, we launched another pouch daemon.
const (
	DaemonLog     = "/tmp/pouchd.log"
	PouchdBin     = "pouchd"
	HomeDir       = "/tmp/test/pouch"
	Listen        = "unix:///tmp/test/pouch/pouchd.sock"
	ContainerdAdd = "/tmp/test/pouch/containerd.sock"
	Pidfile       = "/tmp/test/pouch/pouch.pid"
	ConfigJSON    = "/tmp/pouchconfig.json"
)

// Daemon defines the daemon to test
type Daemon struct {
	daemon.Config

	// pouchd binary location
	Bin string

	// config file
	ConfigJSON string

	// pid of pouchd
	Pid int

	LogPath string
	LogFile *os.File

	// timeout for starting daemon
	timeout int64
}
