package mgr

import (
	"context"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/events"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/kernel"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/system"
	"github.com/alibaba/pouch/registry"
	volumedriver "github.com/alibaba/pouch/storage/volume/driver"
	"github.com/alibaba/pouch/version"

	"github.com/opencontainers/runc/libcontainer/apparmor"
	selinux "github.com/opencontainers/selinux/go-selinux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	unknownHostName      = "<unknown>"
	unknownKernelVersion = "<unknown>"
	unknownOSName        = "<unknown>"
)

//SystemMgr as an interface defines all operations against host.
type SystemMgr interface {
	Info() (types.SystemInfo, error)
	Version() (types.SystemVersion, error)
	Auth(*types.AuthConfig) (string, error)
	UpdateDaemon(*types.DaemonUpdateConfig) error
	SubscribeToEvents(ctx context.Context, since, until time.Time, ef filters.Args) ([]types.EventsMessage, <-chan *types.EventsMessage, <-chan error)
}

// SystemManager is an instance of system management.
type SystemManager struct {
	name     string
	registry *registry.Client
	config   *config.Config
	imageMgr ImageMgr

	store *meta.Store

	eventsService *events.Events
}

// NewSystemManager creates a brand new system manager.
func NewSystemManager(cfg *config.Config, store *meta.Store, imageManager ImageMgr, eventsService *events.Events) (*SystemManager, error) {
	return &SystemManager{
		name:          "system_manager",
		registry:      &registry.Client{},
		config:        cfg,
		imageMgr:      imageManager,
		store:         store,
		eventsService: eventsService,
	}, nil
}

// Info shows system information of daemon.
func (mgr *SystemManager) Info() (types.SystemInfo, error) {
	kernelVersion := unknownKernelVersion
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

	hostname := unknownHostName
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

	OSName := unknownOSName
	if osName, err := system.GetOSName(); err != nil {
		logrus.Warnf("failed to get operating system: %v", err)
	} else {
		OSName = osName
	}

	images, err := mgr.imageMgr.ListImages(context.Background(), filters.NewArgs())
	if err != nil {
		logrus.Warnf("failed to get image info: %v", err)
	}
	volumeDrivers := volumedriver.AllDriversName()

	// security options get four part, seccomp, apparmor, selinux and userns
	securityOpts := []string{}
	sysInfo := system.NewInfo()
	if sysInfo.Seccomp && IsSeccompEnable() {
		securityOpts = append(securityOpts, "seccomp")
	}
	if sysInfo.AppArmor && apparmor.IsEnabled() {
		securityOpts = append(securityOpts, "apparmor")
	}
	if selinux.GetEnabled() {
		securityOpts = append(securityOpts, "selinux")
	}

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
		Driver:            ctrd.CurrentSnapshotterName(context.TODO()),
		// DriverStatus: ,
		ExperimentalBuild: false,
		HTTPProxy:         mgr.config.ImageProxy,
		// HTTPSProxy: ,
		// ID: ,
		CgroupDriver:       mgr.config.GetCgroupDriver(),
		Images:             int64(len(images)),
		IndexServerAddress: "https://index.docker.io/v1/",
		DefaultRegistry:    mgr.config.DefaultRegistry,
		KernelVersion:      kernelVersion,
		Labels:             mgr.config.Labels,
		LiveRestoreEnabled: true,
		LoggingDriver:      mgr.config.DefaultLogConfig.LogDriver,
		VolumeDrivers:      volumeDrivers,
		LxcfsEnabled:       mgr.config.IsLxcfsEnabled,
		CriEnabled:         mgr.config.IsCriEnabled,
		MemTotal:           totalMem,
		Name:               hostname,
		NCPU:               int64(runtime.NumCPU()),
		OperatingSystem:    OSName,
		OSType:             runtime.GOOS,
		PouchRootDir:       mgr.config.HomeDir,
		RegistryConfig:     &mgr.config.RegistryService,
		// RuncCommit: ,
		Runtimes:        mgr.config.Runtimes,
		SecurityOptions: securityOpts,
		ServerVersion:   version.Version,
		ListenAddresses: mgr.config.Listen,
	}
	return info, nil
}

// SubscribeToEvents returns to events on the exchange. Events are sent through the returned
// channel ch. If an error is encountered, it will be sent on channel errs and
// errs will be closed. To end the subscription, cancel the provided context.
func (mgr *SystemManager) SubscribeToEvents(ctx context.Context, since, until time.Time, filter filters.Args) ([]types.EventsMessage, <-chan *types.EventsMessage, <-chan error) {
	ef := events.NewFilter(filter)
	return mgr.eventsService.Subscribe(ctx, since, until, ef)
}

// Version shows version of daemon.
func (mgr *SystemManager) Version() (types.SystemVersion, error) {
	kernelVersion := unknownKernelVersion
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
		GoVersion:     runtime.Version(),
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
		return errors.Wrap(errtypes.ErrInvalidParam, "daemon update config cannot be empty")
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
