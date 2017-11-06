package mgr

import (
	"runtime"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/version"
)

//SystemMgr as an interface defines all operations against host.
type SystemMgr interface {
	Info() (types.SystemInfo, error)
	Version() (types.SystemVersion, error)
}

// SystemManager is an instance of system management.
type SystemManager struct {
	name string
}

// NewSystemManager creates a brand new system manager.
func NewSystemManager(cfg *config.Config) (*SystemManager, error) {
	return &SystemManager{
		name: "system_manager",
	}, nil
}

// Info shows system information of daemon.
func (mgr *SystemManager) Info() (types.SystemInfo, error) {
	return types.SystemInfo{}, nil
}

// Version shows version of daemon.
func (mgr *SystemManager) Version() (types.SystemVersion, error) {
	return types.SystemVersion{
		APIVersion: version.APIVersion,
		Arch:       runtime.GOARCH,
		BuildTime:  version.BuildTime,
		GitCommit:  "",
		GoVersion:  runtime.Version(),
		// TODO:  add a pkg to support getting kernel version
		//KernelVersion: kernel.Version(),
		Os:      runtime.GOOS,
		Version: version.Version,
	}, nil
}
