package mgr

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/kernel"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/system"
	"github.com/alibaba/pouch/registry"
	volumedriver "github.com/alibaba/pouch/storage/volume/driver"
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
	imageMgr ImageMgr

	store *meta.Store
}

// NewSystemManager creates a brand new system manager.
func NewSystemManager(cfg *config.Config, store *meta.Store, imageManager ImageMgr) (*SystemManager, error) {
	return &SystemManager{
		name:     "system_manager",
		registry: &registry.Client{},
		config:   cfg,
		imageMgr: imageManager,
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
		c, ok := obj.(*Container)
		if !ok {
			return nil
		}
		status := c.State.Status
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

	hostname := "<unknown>"
	if name, err := os.Hostname(); err != nil {
		logrus.Warnf("failed to get hostname: %v", err)
	} else {
		hostname = name
	}

	totalMem := int64(0)
	if mem, err := system.GetTotalMem(); err != nil {
		logrus.Warnf("failed to get system mem: %v", err)
	} else {
		totalMem = int64(mem)
	}

	OSName := "<unknown>"
	if osName, err := system.GetOSName(); err != nil {
		logrus.Warnf("failed to get operating system: %v", err)
	} else {
		OSName = osName
	}

	images, err := mgr.imageMgr.ListImages(context.Background(), "")
	if err != nil {
		logrus.Warnf("failed to get image info: %v", err)
	}
	volumeDrivers := volumedriver.AllDriversName()

	info := types.SystemInfo{
		Architecture: runtime.GOARCH,
		// CgroupDriver: ,
		// ContainerdCommit: ,
		Containers:        cRunning + cPaused + cStopped,
		ContainersPaused:  cPaused,
		ContainersRunning: cRunning,
		ContainersStopped: cStopped,
		Debug:             mgr.config.Debug,
		DefaultRuntime:    mgr.config.DefaultRuntime,
		// FIXME: avoid hard code
		Driver: "overlayfs",
		// DriverStatus: ,
		ExperimentalBuild: false,
		HTTPProxy:         mgr.config.ImageProxy,
		// HTTPSProxy: ,
		// ID: ,
		Images:             int64(len(images)),
		IndexServerAddress: "https://index.docker.io/v1/",
		DefaultRegistry:    mgr.config.DefaultRegistry,
		KernelVersion:      kernelVersion,
		Labels:             mgr.config.Labels,
		LiveRestoreEnabled: true,
		LoggingDriver:      mgr.config.DefaultLogConfig.LogDriver,
		VolumeDrivers:      volumeDrivers,
		LxcfsEnabled:       mgr.config.IsLxcfsEnabled,
		MemTotal:           totalMem,
		Name:               hostname,
		NCPU:               int64(runtime.NumCPU()),
		OperatingSystem:    OSName,
		OSType:             runtime.GOOS,
		PouchRootDir:       mgr.config.HomeDir,
		RegistryConfig:     &mgr.config.RegistryService,
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
