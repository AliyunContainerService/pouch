package mgr

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/kernel"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/registry"
	"github.com/alibaba/pouch/version"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//SystemMgr as an interface defines all operations against host.
type SystemMgr interface {
	Info() (types.SystemInfo, error)
	Version() (types.SystemVersion, error)
	Auth(*types.AuthConfig) (string, error)
	UpdateDaemon(*types.DaemonUpdateConfig) error
}

// SystemManager is an instance of system management.
type SystemManager struct {
	name     string
	registry *registry.Client
	config   *config.Config

	store *meta.Store
}

// NewSystemManager creates a brand new system manager.
func NewSystemManager(cfg *config.Config, store *meta.Store) (*SystemManager, error) {
	return &SystemManager{
		name:     "system_manager",
		registry: &registry.Client{},
		config:   cfg,
		store:    store,
	}, nil
}

// Info shows system information of daemon.
func (mgr *SystemManager) Info() (types.SystemInfo, error) {
	kernelVersion := "<unknown>"
	if kv, err := kernel.GetKernelVersion(); err != nil {
		logrus.Warnf("Could not get kernel version: %v", err)
	} else {
		kernelVersion = kv.String()
	}

	var cRunning, cPaused, cStopped int64
	_ = mgr.store.ForEach(func(obj meta.Object) error {
		containerMeta, ok := obj.(*ContainerMeta)
		if !ok {
			return nil
		}
		status := containerMeta.State.Status
		switch status {
		case types.StatusRunning:
			atomic.AddInt64(&cRunning, 1)
		case types.StatusPaused:
			atomic.AddInt64(&cPaused, 1)
		case types.StatusStopped:
			atomic.AddInt64(&cStopped, 1)
		}

		return nil
	})

	info := types.SystemInfo{
		// architecture: ,
		// CgroupDriver: ,
		// ContainerdCommit: ,
		Containers:        cRunning + cPaused + cStopped,
		ContainersPaused:  cPaused,
		ContainersRunning: cRunning,
		ContainersStopped: cStopped,
		Debug:             mgr.config.Debug,
		DefaultRuntime:    mgr.config.DefaultRuntime,
		// Driver: ,
		// DriverStatus: ,
		// ExperimentalBuild: ,
		HTTPProxy: mgr.config.ImageProxy,
		// HTTPSProxy: ,
		// ID: ,
		// Images: ,
		IndexServerAddress: "https://index.docker.io/v1/",
		DefaultRegistry:    mgr.config.DefaultRegistry,
		KernelVersion:      kernelVersion,
		Labels:             mgr.config.Labels,
		// LiveRestoreEnabled: ,
		// LoggingDriver: ,
		// MemTotal: ,
		// Name: ,
		// NCPU: ,
		// OperatingSystem: ,
		OSType:       runtime.GOOS,
		PouchRootDir: mgr.config.HomeDir,
		// RegistryConfig: ,
		// RuncCommit: ,
		// Runtimes: ,
		// SecurityOptions: ,
		ServerVersion:   version.Version,
		ListenAddresses: mgr.config.Listen,
	}
	return info, nil
}

// Version shows version of daemon.
func (mgr *SystemManager) Version() (types.SystemVersion, error) {
	kernelVersion := "<unknown>"
	if kv, err := kernel.GetKernelVersion(); err != nil {
		logrus.Warnf("Could not get kernel version: %v", err)
	} else {
		kernelVersion = kv.String()
	}

	return types.SystemVersion{
		APIVersion:    version.APIVersion,
		Arch:          runtime.GOARCH,
		BuildTime:     version.BuildTime,
		GitCommit:     version.GitCommit,
		GoVersion:     version.GOVersion,
		KernelVersion: kernelVersion,
		Os:            runtime.GOOS,
		Version:       version.Version,
	}, nil
}

// Auth to log in to a registry.
func (mgr *SystemManager) Auth(auth *types.AuthConfig) (string, error) {
	return mgr.registry.Auth(auth)
}

// UpdateDaemon updates config of daemon, only label and image proxy are allowed.
func (mgr *SystemManager) UpdateDaemon(cfg *types.DaemonUpdateConfig) error {
	if cfg == nil || (len(cfg.Labels) == 0 && cfg.ImageProxy == "") {
		return errors.Wrap(errtypes.ErrInvalidParam, fmt.Sprintf("daemon update config cannot be empty"))
	}

	daemonCfg := mgr.config

	daemonCfg.Lock()

	daemonCfg.ImageProxy = cfg.ImageProxy

	length := len(daemonCfg.Labels)
	for _, newLabel := range cfg.Labels {
		appearedKey := false
		newLabelSlice := strings.SplitN(newLabel, "=", 2)
		for i := 0; i < length; i++ {
			oldLabelSlice := strings.SplitN(daemonCfg.Labels[i], "=", 2)
			if newLabelSlice[0] == oldLabelSlice[0] {
				// newLabel's key already appears in daemon's origin labels
				daemonCfg.Labels[i] = newLabel
				appearedKey = true
				continue
			}
		}
		if !appearedKey {
			daemonCfg.Labels = append(daemonCfg.Labels, newLabel)
		}
	}

	daemonCfg.Unlock()

	return nil
}
