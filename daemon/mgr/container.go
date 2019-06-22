package mgr

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/opts"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/daemon/events"
	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/hookplugins"
	"github.com/alibaba/pouch/lxcfs"
	networktypes "github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	mountutils "github.com/alibaba/pouch/pkg/mount"
	"github.com/alibaba/pouch/pkg/streams"
	"github.com/alibaba/pouch/pkg/utils"
	volumetypes "github.com/alibaba/pouch/storage/volume/types"

	"github.com/containerd/cgroups"
	containerdtypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/mount"
	"github.com/docker/go-units"
	"github.com/go-openapi/strfmt"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ContainerMgr as an interface defines all operations against container.
// ContainerMgr's functionality could be divided into three parts:
// 1. regular container management;
// 2. container exec management;
// 3. container network management.
// 4. container copy management
type ContainerMgr interface {
	// 1. the following functions are related to regular container management

	// Load containers from meta store to memory. Split used Restore feature into two function:
	// Load: just load all containers information into memory, it will be called before network
	// manager being initialized.
	// Restore: recover all running containers, it will be called after network manager being
	// initialized so that we can call network functions in the recover procedures.
	Load(ctx context.Context) error

	// Restore recover those alive containers.
	Restore(ctx context.Context) error

	// Create a new container.
	Create(ctx context.Context, name string, config *types.ContainerCreateConfig) (*types.ContainerCreateResp, error)

	// Get the detailed information of container.
	Get(ctx context.Context, name string) (*Container, error)

	// List returns the list of containers.
	List(ctx context.Context, option *ContainerListOption) ([]*Container, error)

	// Kill a running container
	Kill(ctx context.Context, name string, signal uint64) (err error)

	// Start a container.
	Start(ctx context.Context, id string, options *types.ContainerStartOptions) error

	// Stop a container.
	Stop(ctx context.Context, name string, timeout int64) error

	// Restart restart a running container.
	Restart(ctx context.Context, name string, timeout int64) error

	// Pause a container.
	Pause(ctx context.Context, name string) error

	// Unpause a container.
	Unpause(ctx context.Context, name string) error

	// Using a stream to get stats of a container.
	StreamStats(ctx context.Context, name string, config *ContainerStatsConfig) error

	// Stats of a container.
	Stats(ctx context.Context, name string) (*containerdtypes.Metric, *cgroups.Metrics, error)

	// AttachContainerIO attach stream to container IO.
	AttachContainerIO(ctx context.Context, name string, cfg *streams.AttachConfig) error

	// AttachCRILog attach cri log to container IO.
	AttachCRILog(ctx context.Context, name string, path string) error

	// Rename renames a container.
	Rename(ctx context.Context, oldName string, newName string) error

	// Update updates the configurations of a container.
	Update(ctx context.Context, name string, config *types.UpdateConfig) error

	// Upgrade upgrades a container with new image and args.
	Upgrade(ctx context.Context, name string, config *types.ContainerUpgradeConfig) error

	// Top lists the processes running inside of the given container
	Top(ctx context.Context, name string, psArgs string) (*types.ContainerProcessList, error)

	// Resize resizes the size of container tty.
	Resize(ctx context.Context, name string, opts types.ResizeOptions) error

	// Remove removes a container, it may be running or stopped and so on.
	Remove(ctx context.Context, name string, option *types.ContainerRemoveOptions) error

	// Wait stops processing until the given container is stopped.
	Wait(ctx context.Context, name string) (types.ContainerWaitOKBody, error)

	// 2. The following five functions is related to container exec.

	// CreateExec creates exec process's environment.
	CreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (string, error)

	// StartExec executes a new process in container.
	StartExec(ctx context.Context, execid string, cfg *streams.AttachConfig) error

	// InspectExec returns low-level information about exec command.
	InspectExec(ctx context.Context, execid string) (*types.ContainerExecInspect, error)

	// GetExecConfig returns execonfig of a exec process inside container.
	GetExecConfig(ctx context.Context, execid string) (*ContainerExecConfig, error)

	// CheckExecExist check if exec process `name` exist
	CheckExecExist(ctx context.Context, name string) error

	// ResizeExec resizes the size of exec process's tty.
	ResizeExec(ctx context.Context, execid string, opts types.ResizeOptions) error

	// 3. The following two function is related to network management.
	// TODO: inconsistency, Connect/Disconnect operation is in newtork_bridge.go in upper API layer.
	// Here we encapsualted them in container manager, inconsistency exists.

	// Connect is used to connect a container to a network.
	Connect(ctx context.Context, name string, networkIDOrName string, epConfig *types.EndpointSettings) error

	// Disconnect disconnects the given container from
	// given network
	Disconnect(ctx context.Context, containerName, networkName string, force bool) error

	// Logs is used to return log created by the container.
	Logs(ctx context.Context, name string, logsOpt *types.ContainerLogsOptions) (<-chan *logger.LogMessage, bool, error)

	// NewSnapshotsSyncer creates a snapshot syncer.
	NewSnapshotsSyncer(snapshotStore *SnapshotStore, duration time.Duration) *SnapshotsSyncer

	// CreateCheckpoint creates a checkpoint from a running container
	CreateCheckpoint(ctx context.Context, name string, options *types.CheckpointCreateOptions) error

	// ListCheckpoint lists checkpoints from a container
	ListCheckpoint(ctx context.Context, name string, options *types.CheckpointListOptions) ([]string, error)

	// DeleteCheckpoint deletes a checkpoint from a container
	DeleteCheckpoint(ctx context.Context, name string, options *types.CheckpointDeleteOptions) error

	// Commit commits an image from a container.
	Commit(ctx context.Context, name string, options *types.ContainerCommitOptions) (*types.ContainerCommitResp, error)

	// StatPath stats the dir info at the specified path in the container.
	StatPath(ctx context.Context, name, path string) (stat *types.ContainerPathStat, err error)

	// ArchivePath return an archive and dir info at the specified path in the container.
	ArchivePath(ctx context.Context, name, path string) (content io.ReadCloser, stat *types.ContainerPathStat, err error)

	// ExtractToDir extracts the given archive at the specified path in the container.
	ExtractToDir(ctx context.Context, name, path string, copyUIDGID, noOverwriteDirNonDir bool, content io.Reader) error
}

// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
	// Store stores containers in Backend store.
	// Element operated in store must has a type of *Container.
	// By default, Store will use local filesystem with json format to store containers.
	Store *meta.Store

	// Client is used to interact with containerd.
	Client ctrd.APIClient

	// NameToID stores relations between container's name and ID.
	// It is used to get container ID via container name.
	NameToID *collect.SafeMap

	ImageMgr      ImageMgr
	VolumeMgr     VolumeMgr
	NetworkMgr    NetworkMgr
	IOs           *containerio.Cache
	ExecProcesses *collect.SafeMap

	Config *config.Config

	// Cache stores all containers in memory.
	// Element operated in cache must have a type of *Container.
	cache *collect.SafeMap

	// monitor is used to handle container's event, eg: exit, stop and so on.
	monitor *ContainerMonitor

	containerPlugin hookplugins.ContainerPlugin

	// eventsService is used to publish events generated by pouchd
	eventsService *events.Events
}

// NewContainerManager creates a brand new container manager.
func NewContainerManager(ctx context.Context, store *meta.Store, cli ctrd.APIClient, imgMgr ImageMgr, volMgr VolumeMgr, cfg *config.Config, contPlugin hookplugins.ContainerPlugin, eventsService *events.Events) (*ContainerManager, error) {
	mgr := &ContainerManager{
		Store:           store,
		NameToID:        collect.NewSafeMap(),
		Client:          cli,
		ImageMgr:        imgMgr,
		VolumeMgr:       volMgr,
		IOs:             containerio.NewCache(),
		ExecProcesses:   collect.NewSafeMap(),
		cache:           collect.NewSafeMap(),
		Config:          cfg,
		monitor:         NewContainerMonitor(),
		containerPlugin: contPlugin,
		eventsService:   eventsService,
	}

	mgr.Client.SetExitHooks(mgr.exitedAndRelease)
	mgr.Client.SetExecExitHooks(mgr.execExitedAndRelease)
	mgr.Client.SetEventsHooks(mgr.publishContainerdEvent, mgr.updateContainerState)

	go mgr.execProcessGC()

	return mgr, nil
}

// Load containers from meta store to memory.
func (mgr *ContainerManager) Load(ctx context.Context) error {
	fn := func(obj meta.Object) error {
		container, ok := obj.(*Container)
		if !ok {
			// object has not type of Container
			return nil
		}

		id := container.ID
		// map container's name to id.
		mgr.NameToID.Put(container.Name, id)

		// put container into cache.
		mgr.cache.Put(id, container)

		return nil
	}

	return mgr.Store.ForEach(fn)
}

// Restore tries to recover those alive containers
func (mgr *ContainerManager) Restore(ctx context.Context) error {
	// get all running containers
	containers, err := mgr.List(ctx,
		&ContainerListOption{
			All: true,
		},
	)
	if err != nil {
		logrus.Errorf("failed to get container list when restore containers: %v", err)
		return errors.Wrap(err, "failed to get container list")
	}

	for _, c := range containers {
		id := c.Key()

		// NOTE: when pouch is restarting, we need to initialize
		// container IO for the existing containers just in case that
		// user tries to restart the stopped containers.
		cntrio, err := mgr.initContainerIO(c)
		if err != nil {
			logrus.Errorf("failed to init container IO %s: %v", id, err)
			return err
		}

		if err := mgr.initLogDriverBeforeStart(c); err != nil {
			logrus.Errorf("failed to init log driver %s: %v", id, err)
			return err
		}

		// recover the running or paused container.
		if !c.IsRunningOrPaused() {
			continue
		}

		logrus.Debugf("Start recover container %s", id)

		// Start recover the container
		err = mgr.Client.RecoverContainer(ctx, id, cntrio)
		if err == nil {
			continue
		}

		// Note(ziren): Since we got an unknown error when recover the
		// container, we just log the error and continue in case we wrongly
		// release the container's resources
		if !strings.Contains(err.Error(), "not found") {
			logrus.Errorf("failed to recover container %s: %v", id, err)
			// release io
			cntrio.Close()
			mgr.IOs.Remove(id)
			continue
		}

		// Note(ziren) if containerd post not found error, that is mean
		// container or task is not found. So we should set the container's
		// status to exited and release the container's resources.
		logrus.Warnf("recover container %s, got a notfound error, start clean the container's resources", id)
		if err := mgr.exitedAndRelease(id, nil, nil); err != nil {
			logrus.Errorf("failed to execute exited and release for container %s: %v", id, err)
		}
	}

	return nil
}

// Create checks passed in parameters and create a Container object whose status is set at Created.
func (mgr *ContainerManager) Create(ctx context.Context, name string, config *types.ContainerCreateConfig) (resp *types.ContainerCreateResp, err error) {
	currentSnapshotter := ctrd.CurrentSnapshotterName(ctx)
	config.Snapshotter = currentSnapshotter

	if mgr.containerPlugin != nil {
		logrus.Infof("invoke container pre-create hook in plugin")
		if ex := mgr.containerPlugin.PreCreate(config); ex != nil {
			return nil, errors.Wrapf(ex, "pre-create plugin point execute failed")
		}
	}

	// Attention, since we support multi snapshotter, if snapshotter not changed,
	// means plugin not change it, so remove value in case to effect origin logic
	if config.Snapshotter == currentSnapshotter {
		config.Snapshotter = ""
	}

	// NOTE: choose snapshotter, snapshotter can only be set
	// through containerPlugin in Create function
	ctx = ctrd.WithSnapshotter(ctx, config.Snapshotter)

	// cleanup allocated resources when failed
	cleanups := []func() error{}
	defer func() {
		// do cleanup
		if err != nil {
			logrus.Infof("start to rollback allocated resources of container %v", name)
			for _, f := range cleanups {
				nerr := f()
				if nerr != nil {
					logrus.Errorf("fail to cleanup allocated resource, error is %v", nerr)
				}
			}
		}
	}()

	imgID, _, primaryRef, err := mgr.ImageMgr.CheckReference(ctx, config.Image)
	if err != nil {
		return nil, err
	}
	config.Image = primaryRef.String()

	// TODO: check request validate.
	if config.HostConfig == nil {
		return nil, errors.Wrapf(errtypes.ErrInvalidParam, "HostConfig cannot be empty")
	}
	if config.NetworkingConfig == nil {
		return nil, errors.Wrapf(errtypes.ErrInvalidParam, "NetworkingConfig cannot be empty")
	}

	// validate disk quota
	if err := mgr.validateDiskQuota(config); err != nil {
		return nil, errors.Wrapf(err, "invalid disk quota config")
	}

	id, err := mgr.generateContainerID(config.SpecificID)
	if err != nil {
		return nil, err
	}
	//put container id to cache to prevent concurrent containerCreateReq with same specific id
	mgr.cache.Put(id, nil)
	defer func() {
		//clear cache
		if err != nil {
			mgr.cache.Remove(id)
		}
	}()

	if name == "" {
		name = mgr.generateName(id)
	} else if mgr.NameToID.Get(name).Exist() {
		return nil, errors.Wrapf(errtypes.ErrAlreadyExisted, "container name %s", name)
	}

	// set hostname.
	if config.Hostname.String() == "" {
		// if hostname is empty, take the part of id as the hostname
		config.Hostname = strfmt.Hostname(id[:12])
	}

	// set container runtime
	if config.HostConfig.Runtime == "" {
		config.HostConfig.Runtime = mgr.Config.DefaultRuntime
	}

	config.HostConfig.RuntimeType, err = mgr.getRuntimeType(config.HostConfig.Runtime)
	if err != nil {
		return nil, errors.Wrapf(errtypes.ErrInvalidParam, "unknown runtime %s: %v", config.HostConfig.Runtime, err)
	}

	snapID := id
	// create a snapshot with image.
	if err := mgr.Client.CreateSnapshot(ctx, snapID, config.Image); err != nil {
		return nil, err
	}
	cleanups = append(cleanups, func() error {
		logrus.Infof("start to cleanup snapshot, id is %v", id)
		return mgr.Client.RemoveSnapshot(ctx, id)
	})

	// set lxcfs binds
	if config.HostConfig.EnableLxcfs && lxcfs.IsLxcfsEnabled {
		config.HostConfig.Binds = append(config.HostConfig.Binds, lxcfs.LxcfsParentDir+":/var/lib/lxc:shared")
		sourceDir := lxcfs.LxcfsHomeDir + "/proc/"
		destDir := "/proc/"
		for _, procFile := range lxcfs.LxcfsProcFiles {
			bind := fmt.Sprintf("%s%s:%s%s", sourceDir, procFile, destDir, procFile)
			config.HostConfig.Binds = append(config.HostConfig.Binds, bind)
		}
	}

	// set default log driver and validate for logger driver
	config.HostConfig.LogConfig = mgr.getDefaultLogConfigIfMissing(config.HostConfig.LogConfig)

	// set ReadonlyPaths and MaskedPaths to nil if privileged was set.
	if config.HostConfig.Privileged {
		config.HostConfig.ReadonlyPaths = nil
		config.HostConfig.MaskedPaths = nil
	}

	container := &Container{
		State: &types.ContainerState{
			Status:     types.StatusCreated,
			StartedAt:  time.Time{}.UTC().Format(utils.TimeLayout),
			FinishedAt: time.Time{}.UTC().Format(utils.TimeLayout),
		},
		ID:         id,
		Image:      imgID.String(),
		Name:       name,
		Config:     &config.ContainerConfig,
		Created:    time.Now().UTC().Format(utils.TimeLayout),
		HostConfig: config.HostConfig,
		SnapshotID: snapID,
	}

	if _, err := mgr.initContainerIO(container); err != nil {
		logrus.Errorf("failed to initialise IO: %v", err)
		return nil, err
	}

	// merge image's config into container
	if err := container.merge(func() (ocispec.ImageConfig, error) {
		return mgr.ImageMgr.GetOCIImageConfig(ctx, config.Image)
	}); err != nil {
		return nil, err
	}

	// set container basefs, basefs is not created in pouchd, it will created
	// after create options passed to containerd.
	mgr.setBaseFS(ctx, container)

	// init container storage module, such as: set volumes, set diskquota, set /etc/mtab, copy image's data to volume.
	if err := mgr.initContainerStorage(ctx, container); err != nil {
		return nil, errors.Wrapf(err, "failed to init container storage, id: (%s)", container.ID)
	}

	// set network settings
	networkMode := config.HostConfig.NetworkMode
	if networkMode == "" {
		config.HostConfig.NetworkMode = "bridge"
	}
	container.NetworkSettings = new(types.NetworkSettings)
	if len(config.NetworkingConfig.EndpointsConfig) > 0 {
		container.NetworkSettings.Networks = config.NetworkingConfig.EndpointsConfig
	}
	if container.NetworkSettings.Networks == nil &&
		!IsContainer(config.HostConfig.NetworkMode) && !IsNetNS(config.HostConfig.NetworkMode) {
		container.NetworkSettings.Networks = make(map[string]*types.EndpointSettings)
		container.NetworkSettings.Networks[config.HostConfig.NetworkMode] = new(types.EndpointSettings)
	}
	container.NetworkSettings.Ports = config.HostConfig.PortBindings

	if err := parseSecurityOpts(container, config.HostConfig.SecurityOpt); err != nil {
		return nil, err
	}

	// Get snapshot UpperDir
	mounts, err := mgr.Client.GetMounts(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(mounts) != 1 {
		return nil, fmt.Errorf("failed to get snapshot %s mounts: not equals one", id)
	}
	container.SetSnapshotterMeta(mounts)

	// amendContainerSettings modify container config settings to wanted
	amendContainerSettings(&config.ContainerConfig, config.HostConfig)

	// validate container Config
	warnings, err := mgr.validateConfig(container, false)
	if err != nil {
		return nil, err
	}

	// store disk
	if err := container.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update meta: %v", err)
		return nil, err
	}

	// add to collection
	mgr.NameToID.Put(name, id)
	mgr.cache.Put(id, container)

	mgr.LogContainerEvent(ctx, container, "create")

	return &types.ContainerCreateResp{
		ID:       id,
		Name:     name,
		Warnings: warnings,
	}, nil
}

func (mgr *ContainerManager) getDefaultLogConfigIfMissing(logConfig *types.LogConfig) *types.LogConfig {
	defaultLogOpts := make(map[string]string)
	for k, v := range mgr.Config.DefaultLogConfig.LogOpts {
		defaultLogOpts[k] = v
	}

	if logConfig == nil {
		defaultConfig := mgr.Config.DefaultLogConfig
		defaultConfig.LogOpts = defaultLogOpts
		return &defaultConfig
	}

	if logConfig.LogDriver == "" {
		logConfig.LogDriver = mgr.Config.DefaultLogConfig.LogDriver
	}

	if len(logConfig.LogOpts) == 0 {
		logConfig.LogOpts = defaultLogOpts
	}

	return logConfig
}

// Get the detailed information of container.
func (mgr *ContainerManager) Get(ctx context.Context, name string) (*Container, error) {
	c, err := mgr.container(name)
	if err != nil {
		return nil, err
	}
	cID := c.ID

	// get all execids belongs to this container
	fn := func(v interface{}) bool {
		execConfig, ok := v.(*ContainerExecConfig)
		if !ok || execConfig.ContainerID != cID {
			return false
		}

		return true
	}

	var execIDs []string
	execProcesses := mgr.ExecProcesses.Values(fn)
	for k := range execProcesses {
		execIDs = append(execIDs, k)
	}
	c.ExecIds = execIDs

	return c, nil
}

// Start a pre created Container.
func (mgr *ContainerManager) Start(ctx context.Context, id string, options *types.ContainerStartOptions) (err error) {
	if id == "" {
		return errors.Wrap(errtypes.ErrInvalidParam, "container ID cannot empty")
	}

	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	// NOTE: choose snapshotter, snapshotter can only be set
	// through containerPlugin in Create function
	ctx = ctrd.WithSnapshotter(ctx, c.Config.Snapshotter)

	err = mgr.start(ctx, c, options)
	if err == nil {
		mgr.LogContainerEvent(ctx, c, "start")
	}

	return err
}

func (mgr *ContainerManager) start(ctx context.Context, c *Container, options *types.ContainerStartOptions) error {
	// NOTE: add a big lock when start a container
	c.Lock()
	defer c.Unlock()

	var err error
	c.DetachKeys = options.DetachKeys

	// check if container's status is paused
	if c.State.Paused {
		return fmt.Errorf("cannot start a paused container, try unpause instead")
	}

	// check if container's status is running
	if c.State.Running {
		return errors.Wrapf(errtypes.ErrNotModified, "container already started")
	}

	if c.State.Dead {
		return fmt.Errorf("cannot start a dead container %s", c.ID)
	}

	attachedVolumes := map[string]struct{}{}
	defer func() {
		if err == nil {
			return
		}

		// release the container resources(network and containerio)
		err = mgr.releaseContainerResources(c)
		if err != nil {
			logrus.Errorf("failed to release container(%s) resources: %v", c.ID, err)
		}

		// detach the volumes
		for name := range attachedVolumes {
			if _, err = mgr.VolumeMgr.Detach(ctx, name, map[string]string{volumetypes.OptionRef: c.ID}); err != nil {
				logrus.Errorf("failed to detach volume(%s) when start container(%s) rollback: %v", name, c.ID, err)
			}
		}
	}()

	for _, mp := range c.Mounts {
		if mp.Name == "" {
			continue
		}
		if _, err = mgr.VolumeMgr.Attach(ctx, mp.Name, map[string]string{volumetypes.OptionRef: c.ID}); err != nil {
			return errors.Wrapf(err, "failed to attach volume(%s)", mp.Name)
		}
		attachedVolumes[mp.Name] = struct{}{}
	}

	if err = mgr.prepareContainerNetwork(ctx, c); err != nil {
		return err
	}

	if err = mgr.createContainerdContainer(ctx, c, options.CheckpointDir, options.CheckpointID); err != nil {
		return errors.Wrapf(err, "failed to create container(%s) on containerd", c.ID)
	}

	return nil
}

func (mgr *ContainerManager) prepareContainerNetwork(ctx context.Context, c *Container) error {
	networkMode := c.HostConfig.NetworkMode

	if IsContainer(networkMode) {
		var origContainer *Container
		origContainer, err := mgr.Get(ctx, strings.SplitN(networkMode, ":", 2)[1])
		if err != nil {
			return err
		}

		c.HostnamePath = origContainer.HostnamePath
		c.HostsPath = origContainer.HostsPath
		c.ResolvConfPath = origContainer.ResolvConfPath
		c.Config.Hostname = origContainer.Config.Hostname
		c.Config.Domainname = origContainer.Config.Domainname

		return nil
	}

	// initialise host network mode
	if IsHost(networkMode) {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		c.Config.Hostname = strfmt.Hostname(hostname)
	}

	// build the network related path.
	if err := mgr.buildNetworkRelatedPath(c); err != nil {
		return err
	}

	// network is prepared by upper system. do nothing here.
	if IsNetNS(networkMode) {
		return nil
	}

	// initialise network endpoint
	if c.NetworkSettings == nil {
		return nil
	}

	for name, endpointSetting := range c.NetworkSettings.Networks {
		endpoint := mgr.buildContainerEndpoint(c, name)
		endpoint.EndpointConfig = endpointSetting
		if _, err := mgr.NetworkMgr.EndpointCreate(ctx, endpoint); err != nil {
			logrus.Errorf("failed to create endpoint: %v", err)
			return err
		}
	}

	sb, err := mgr.NetworkMgr.Controller().SandboxByID(c.NetworkSettings.SandboxID)
	if err != nil {
		// sandbox not found, maybe caused by disconnect network or no endpoint
		logrus.Warnf("failed to get sandbox by id(%s), err(%v)", c.NetworkSettings.SandboxID, err)
		c.NetworkSettings.Ports = types.PortMap{}
		return nil
	}

	c.NetworkSettings.Ports = getSandboxPortMapInfo(sb)
	return nil
}

// buildNetworkRelatedPath builds the network related path.
func (mgr *ContainerManager) buildNetworkRelatedPath(c *Container) error {
	// set the hosts file path.
	c.HostsPath = path.Join(mgr.Store.Path(c.ID), "hosts")

	// set the resolv.conf file path.
	c.ResolvConfPath = path.Join(mgr.Store.Path(c.ID), "resolv.conf")

	// set the hostname file path.
	c.HostnamePath = path.Join(mgr.Store.Path(c.ID), "hostname")

	// write the hostname file, other files are filled by libnetwork.
	return ioutil.WriteFile(c.HostnamePath, []byte(c.Config.Hostname+"\n"), 0644)
}

func (mgr *ContainerManager) createContainerdContainer(ctx context.Context, c *Container, checkpointDir, checkpointID string) error {
	// CgroupParent from HostConfig will be first priority to use,
	// then will be value from mgr.Config.CgroupParent
	if c.HostConfig.CgroupParent == "" {
		c.HostConfig.CgroupParent = mgr.Config.CgroupParent
	}

	var (
		err     error
		prioArr []int
		argsArr [][]string
	)

	// if creating the container by specified the rootfs, we must check
	// whether the rootfs is mounted before creation
	if c.RootFSProvided {
		if err := mgr.ensureRootFSMounted(c.BaseFS, c.Snapshotter.Data); err != nil {
			return fmt.Errorf("failed to mount container rootfs: %v", err)
		}
	}

	if mgr.containerPlugin != nil {
		// TODO: make func PreStart with no data race
		prioArr, argsArr, err = mgr.containerPlugin.PreStart(c)
		if err != nil {
			return errors.Wrapf(err, "get pre-start hook error from container plugin")
		}
	}

	sw := &SpecWrapper{
		ctrMgr:     mgr,
		volMgr:     mgr.VolumeMgr,
		netMgr:     mgr.NetworkMgr,
		prioArr:    prioArr,
		argsArr:    argsArr,
		useSystemd: mgr.Config.UseSystemd(),
	}

	if err = createSpec(ctx, c, sw); err != nil {
		return err
	}

	// init log driver
	if err := mgr.initLogDriverBeforeStart(c); err != nil {
		return errors.Wrap(err, "failed to initialize log driver")
	}

	// set container's LogPath
	mgr.SetContainerLogPath(c)

	if c.HostConfig.RuntimeType == "" {
		c.HostConfig.RuntimeType = ctrd.RuntimeTypeV1
	}

	runtimeOptions, err := mgr.generateRuntimeOptions(c.HostConfig.Runtime)
	if err != nil {
		return err
	}

	ctrdContainer := &ctrd.Container{
		ID:             c.ID,
		Image:          c.Config.Image,
		Labels:         c.Config.Labels,
		RuntimeType:    c.HostConfig.RuntimeType,
		RuntimeOptions: runtimeOptions,
		Spec:           sw.s,
		IO:             mgr.IOs.Get(c.ID),
		RootFSProvided: c.RootFSProvided,
		BaseFS:         c.BaseFS,
		UseSystemd:     mgr.Config.UseSystemd(),
	}
	// make sure the SnapshotID got a proper value
	ctrdContainer.SnapshotID = c.SnapshotKey()

	if checkpointID != "" {
		checkpointDir, err = mgr.getCheckpointDir(c.ID, checkpointDir, checkpointID, false)
		if err != nil {
			return err
		}
	}
	if err := mgr.Client.CreateContainer(ctx, ctrdContainer, checkpointDir); err != nil {
		logrus.Errorf("failed to create new containerd container: %v", err)

		// TODO(ziren): markStoppedAndRelease may failed
		// we should clean resources of container when start failed
		_ = mgr.markStoppedAndRelease(c, nil)
		return err
	}

	// Create containerd container success.

	pid, err := mgr.Client.ContainerPID(ctx, c.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to get PID of container %s", c.ID)
	}

	c.SetStatusRunning(int64(pid))

	// set Snapshot MergedDir
	c.Snapshotter.Data["MergedDir"] = c.BaseFS

	return c.Write(mgr.Store)
}

func (mgr *ContainerManager) ensureRootFSMounted(rootfs string, snapData map[string]string) error {
	if rootfs == "" || len(snapData) == 0 {
		return fmt.Errorf("container rootfs or snapshotter data is empty")
	}

	// check if rootfs already mounted
	notMounted, err := mountutils.IsLikelyNotMountPoint(rootfs)
	if err != nil {
		return err
	}
	// rootfs already mounted
	if !notMounted {
		return nil
	}

	var workDir, upperDir, lowerDir string
	for _, dir := range []string{"WorkDir", "UpperDir", "LowerDir"} {
		if v, ok := snapData[dir]; ok {
			switch dir {
			case "WorkDir":
				workDir = v
			case "UpperDir":
				upperDir = v
			case "LowerDir":
				lowerDir = v
			}
		}
	}

	if workDir == "" || upperDir == "" || lowerDir == "" {
		return fmt.Errorf("faile to mount overlay: one or more dirs in WorkDir, UpperDir and LowerDir are empty")
	}

	options := []string{
		fmt.Sprintf("workdir=%s", snapData["WorkDir"]),
		fmt.Sprintf("upperdir=%s", snapData["UpperDir"]),
		fmt.Sprintf("lowerdir=%s", snapData["LowerDir"]),
	}
	mount := mount.Mount{
		Type:    "overlay",
		Source:  "overlay",
		Options: options,
	}

	return mount.Mount(rootfs)
}

// Stop stops a running container.
func (mgr *ContainerManager) Stop(ctx context.Context, name string, timeout int64) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	// NOTE: choose snapshotter, snapshotter can only be set
	// through containerPlugin in Create function
	ctx = ctrd.WithSnapshotter(ctx, c.Config.Snapshotter)

	err = mgr.stop(ctx, c, timeout)
	if err != nil {
		return err
	}
	mgr.LogContainerEvent(ctx, c, "stop")

	return nil
}

func (mgr *ContainerManager) stop(ctx context.Context, c *Container, timeout int64) error {
	c.Lock()
	defer c.Unlock()

	if !c.IsRunningOrPaused() {
		// stopping a non-running container is valid.
		return nil
	}

	if timeout == 0 {
		timeout = c.StopTimeout()
	}

	id := c.ID
	msg, err := mgr.Client.DestroyContainer(ctx, id, timeout)
	if err != nil {
		return errors.Wrapf(err, "failed to destroy container %s", id)
	}

	return mgr.markStoppedAndRelease(c, msg)
}

// Restart restarts a running container.
func (mgr *ContainerManager) Restart(ctx context.Context, name string, timeout int64) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	// NOTE: choose snapshotter, snapshotter can only be set
	// through containerPlugin in Create function
	ctx = ctrd.WithSnapshotter(ctx, c.Config.Snapshotter)

	if c.IsRunningOrPaused() {
		// stop container if it is running or paused.
		if err := mgr.stop(ctx, c, timeout); err != nil {
			ex := fmt.Errorf("failed to stop container %s when restarting: %v", c.ID, err)
			logrus.Errorf(ex.Error())
			return ex
		}
	}

	logrus.Debugf("start container %s when restarting", c.ID)

	// start container
	err = mgr.start(ctx, c, &types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// count start times
	c.RestartCount++

	logrus.Debugf("container %s restartCount is %d", c.ID, c.RestartCount)
	return c.Write(mgr.Store)
}

// Pause pauses a running container.
func (mgr *ContainerManager) Pause(ctx context.Context, name string) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if !c.State.Running {
		return fmt.Errorf("container's status is not running: %s", c.State.Status)
	}

	if err := mgr.Client.PauseContainer(ctx, c.ID); err != nil {
		return errors.Wrapf(err, "failed to pause container %s", c.ID)
	}

	c.SetStatusPaused()

	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update meta of container %s: %v", c.ID, err)
		return err
	}
	mgr.LogContainerEvent(ctx, c, "pause")

	return nil
}

// Unpause unpauses a paused container.
func (mgr *ContainerManager) Unpause(ctx context.Context, name string) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if !c.State.Paused {
		return fmt.Errorf("status(%s) of container %s is not paused", c.State.Status, c.ID)
	}

	if err := mgr.Client.UnpauseContainer(ctx, c.ID); err != nil {
		return errors.Wrapf(err, "failed to unpause container %s", c.ID)
	}

	c.SetStatusUnpaused()

	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update meta of container %s: %v", c.ID, err)
		return err
	}

	mgr.LogContainerEvent(ctx, c, "unpause")
	return nil
}

// AttachContainerIO attachs a container's io.
func (mgr *ContainerManager) AttachContainerIO(ctx context.Context, name string, cfg *streams.AttachConfig) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	cntrio := mgr.IOs.Get(c.ID)
	cfg.Terminal = c.Config.Tty

	// NOTE: the AttachContainerIO might use the hijack's connection as
	// stdin in the AttachConfig. If we close it directly, the stdout/stderr
	// will return the `using closed connection` error. As a result, the
	// Attach will return the error. We need to use pipe here instead of
	// origin one and let the caller closes the stdin by themself.
	if c.Config.OpenStdin && cfg.UseStdin {
		oldStdin := cfg.Stdin
		pstdinr, pstdinw := io.Pipe()
		go func() {
			defer pstdinw.Close()
			io.Copy(pstdinw, oldStdin)
		}()
		cfg.Stdin = pstdinr
		cfg.CloseStdin = true
	} else {
		cfg.UseStdin = false
	}
	return <-cntrio.Stream().Attach(ctx, cfg)
}

// AttachCRILog adds cri log to a container.
func (mgr *ContainerManager) AttachCRILog(ctx context.Context, name string, logPath string) error {
	if logPath == "" {
		return errors.Wrap(errtypes.ErrInvalidParam, "logPath cannot be empty")
	}
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	return mgr.attachCRILog(c, logPath)
}

// Rename renames a container.
func (mgr *ContainerManager) Rename(ctx context.Context, oldName, newName string) error {
	if mgr.NameToID.Get(newName).Exist() {
		return errors.Wrapf(errtypes.ErrAlreadyExisted, "container name %s", newName)
	}

	c, err := mgr.container(oldName)
	if err != nil {
		return errors.Wrapf(err, "failed to rename container %s", oldName)
	}

	c.Lock()
	defer c.Unlock()

	if c.State.Dead {
		return fmt.Errorf("cannot rename a dead container %s", c.ID)
	}

	attributes := map[string]string{
		"oldName": oldName,
	}

	name := c.Name
	c.Name = newName

	mgr.NameToID.Remove(name)
	mgr.NameToID.Put(newName, c.ID)

	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update meta of container %s: %v", c.ID, err)
		return err
	}

	mgr.LogContainerEventWithAttributes(ctx, c, "rename", attributes)
	return nil
}

// Update updates the configurations of a container.
func (mgr *ContainerManager) Update(ctx context.Context, name string, config *types.UpdateConfig) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	warnings, err := validateResource(&config.Resources, true)
	if err != nil {
		return err
	}
	if len(warnings) != 0 {
		logrus.Warnf("warnings update %s: %v", name, warnings)
	}

	restore := false
	oldConfig := *c.Config
	oldHostconfig := *c.HostConfig
	defer func() {
		if restore {
			c.Config = &oldConfig
			c.HostConfig = &oldHostconfig
		}
	}()

	if c.State.Running && config.Resources.KernelMemory != 0 {
		return fmt.Errorf("failed to update container %s: can not update kernel memory to a running container, please stop it first", c.ID)
	}

	if c.State.Dead {
		return fmt.Errorf("cannot update a dead container %s", c.ID)
	}

	// update container disk quota
	if err := mgr.updateContainerDiskQuota(ctx, c, config.DiskQuota); err != nil {
		return errors.Wrapf(err, "failed to update diskquota of container %s", c.ID)
	}

	// init Container Labels
	if c.Config.Labels == nil {
		c.Config.Labels = map[string]string{}
	}

	// compatibility with alidocker, UpdateConfig.Label is []string
	// but ContainerConfig.Labels is map[string]string
	if len(config.Label) != 0 {
		// support remove some labels
		newLabels := opts.ParseLabels(config.Label)

		for k, v := range newLabels {
			if v == "" {
				delete(c.Config.Labels, k)
			} else {
				c.Config.Labels[k] = v
			}
		}
	}

	// update Resources of a container.
	if err := mgr.updateContainerResources(c, config.Resources); err != nil {
		restore = true
		return errors.Wrapf(err, "failed to update resource of container %s", c.ID)
	}

	// TODO update restartpolicy when container is running.
	if config.RestartPolicy != nil && config.RestartPolicy.Name != "" {
		c.HostConfig.RestartPolicy = config.RestartPolicy
	}

	// Update Env
	newEnvSlice, err := mergeEnvSlice(config.Env, c.Config.Env)
	if err != nil {
		return err
	}
	c.Config.Env = newEnvSlice

	// update env when container is running, default snapshotter driver
	// is overlayfs
	if c.IsRunningOrPaused() && len(config.Env) > 0 && c.Snapshotter != nil {
		if mergedDir, exists := c.Snapshotter.Data["MergedDir"]; exists {
			if err := updateContainerEnv(c.Config.Env, mergedDir); err != nil {
				return errors.Wrapf(err, "failed to update env of running container")
			}
		}
	}

	if len(config.SpecAnnotation) > 0 {
		c.Config.SpecAnnotation = mergeAnnotation(config.SpecAnnotation, c.Config.SpecAnnotation)
	}

	if mgr.containerPlugin != nil && len(config.Env) > 0 {
		if err = mgr.containerPlugin.PostUpdate(c.BaseFS, c.Config.Env); err != nil {
			return err
		}
	}

	// If container is not running, update container metadata struct is enough,
	// resources will be updated when the container is started again,
	// If container is running, we need to update configs to the real world.
	if c.State.Running {
		if err := mgr.Client.UpdateResources(ctx, c.ID, c.HostConfig.Resources); err != nil {
			restore = true
			return fmt.Errorf("failed to update resource: %s", err)
		}
	}

	// store disk.
	err = c.Write(mgr.Store)
	if err != nil {
		restore = true
	}

	mgr.LogContainerEvent(ctx, c, "update")
	return err
}

// Remove removes a container, it may be running or stopped and so on.
func (mgr *ContainerManager) Remove(ctx context.Context, name string, options *types.ContainerRemoveOptions) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	// NOTE: choose snapshotter, snapshotter can only be set
	// through containerPlugin in Create function
	ctx = ctrd.WithSnapshotter(ctx, c.Config.Snapshotter)

	c.Lock()
	defer c.Unlock()

	if c.IsRunningOrPaused() && !options.Force {
		return fmt.Errorf("container %s is not stopped, cannot remove it without flag force", c.ID)
	}

	if c.State.Dead {
		logrus.Warnf("container has been deleted %s", c.ID)
		return nil
	}

	// if the container is running, force to stop it.
	if c.IsRunningOrPaused() && options.Force {
		_, err := mgr.Client.DestroyContainer(ctx, c.ID, c.StopTimeout())
		if err != nil && !errtypes.IsNotfound(err) {
			return errors.Wrapf(err, "failed to destroy container %s when removing", c.ID)
		}
		// After stopping a running container, we should release container resource
		c.UnsetMergedDir()
		if err := mgr.releaseContainerResources(c); err != nil {
			logrus.Errorf("failed to release container %s resources when removing: %v", c.ID, err)
		}
	}

	if err := mgr.detachVolumes(ctx, c, options.Volumes); err != nil {
		logrus.Errorf("failed to detach volume: %v", err)
	}

	// if creating the container by specify rootfs,
	// we should umount the rootfs when delete the container.
	if c.RootFSProvided {
		if err := mount.Unmount(c.BaseFS, 0); err != nil {
			logrus.Errorf("failed to umount rootfs when remove the container %s: %v", c.ID, err)
		}

		// Note(ziren): when deleting a container whose rootfs was provided, we also should
		// remove the upperDir and workDir of container. because the directories cost disk
		// space and the disk space counted into the new container that using the same
		// disk quota id.
		if err := c.CleanRootfsSnapshotDirs(); err != nil {
			logrus.Errorf("failed to clean rootfs: %v", err)
		}
	} else if err := mgr.Client.RemoveSnapshot(ctx, c.SnapshotKey()); err != nil {
		// if the container is created by normal method, remove the
		// snapshot when delete it.
		logrus.Errorf("failed to remove snapshot of container %s: %v", c.ID, err)
	}

	// When removing a container, we have set up such rule for object removing sequences:
	// 1. container object in pouchd's memory;
	// 2. meta.json for container in local disk.
	// 3. remove the container IO from cache

	// remove name
	mgr.NameToID.Remove(c.Name)
	// remove container cache
	mgr.cache.Remove(c.ID)
	// remove the container IO
	mgr.IOs.Remove(c.ID)
	c.State.Dead = true

	logRootDir, err := mgr.getLogRootDirFromOpt(c, false)
	if err == nil && logRootDir != mgr.Store.Path(c.ID) {
		rErr := os.RemoveAll(logRootDir)
		if rErr != nil {
			logrus.Warnf("failed to remove container %s log path %s: %v", c.ID, logRootDir, rErr)
		}
	}

	// remove meta.json for container in local disk
	if err := mgr.Store.Remove(c.Key()); err != nil {
		logrus.Errorf("failed to remove container %s from meta store: %v", c.ID, err)
	}

	mgr.LogContainerEvent(ctx, c, "destroy")
	return nil
}

func (mgr *ContainerManager) updateContainerDiskQuota(ctx context.Context, c *Container, diskQuota map[string]string) (err error) {
	if diskQuota == nil {
		return nil
	}

	// backup diskquota
	origDiskQuota := c.Config.DiskQuota
	defer func() {
		if err != nil {
			c.Config.DiskQuota = origDiskQuota
		}
	}()

	if c.Config.DiskQuota == nil {
		c.Config.DiskQuota = make(map[string]string)
	}
	for dir, quota := range diskQuota {
		c.Config.DiskQuota[dir] = quota
	}

	// set mount point disk quota
	if err = mgr.setDiskQuota(ctx, c, false); err != nil {
		return errors.Wrapf(err, "failed to set mount point disk quota")
	}

	return nil
}

// updateContainerResources update container's resources parameters.
func (mgr *ContainerManager) updateContainerResources(c *Container, resources types.Resources) error {
	// update resources of container.
	cResources := &c.HostConfig.Resources
	if resources.BlkioWeight != 0 {
		cResources.BlkioWeight = resources.BlkioWeight
	}
	if len(resources.BlkioDeviceReadBps) != 0 {
		cResources.BlkioDeviceReadBps = resources.BlkioDeviceReadBps
	}
	if len(resources.BlkioDeviceReadIOps) != 0 {
		cResources.BlkioDeviceReadIOps = resources.BlkioDeviceReadIOps
	}
	if len(resources.BlkioDeviceWriteBps) != 0 {
		cResources.BlkioDeviceWriteBps = resources.BlkioDeviceWriteBps
	}
	if len(resources.BlkioDeviceWriteIOps) != 0 {
		cResources.BlkioDeviceWriteIOps = resources.BlkioDeviceWriteIOps
	}
	if resources.CPUPeriod != 0 {
		cResources.CPUPeriod = resources.CPUPeriod
	}
	if resources.CPUQuota == -1 || resources.CPUQuota >= 1000 {
		cResources.CPUQuota = resources.CPUQuota
	}
	if resources.CPUShares != 0 {
		cResources.CPUShares = resources.CPUShares
	}
	if resources.CpusetCpus != "" {
		cResources.CpusetCpus = resources.CpusetCpus
	}
	if resources.CpusetMems != "" {
		cResources.CpusetMems = resources.CpusetMems
	}
	if resources.Memory != 0 {
		// if memory limit smaller than already set memoryswap limit and doesn't
		// update the memoryswap limit, then error out.
		if cResources.MemorySwap != 0 && resources.Memory > cResources.MemorySwap && resources.MemorySwap == 0 {
			return fmt.Errorf("Memory limit should be smaller than already set memoryswap limit, update the memoryswap at the same time")
		}
		cResources.Memory = resources.Memory
	}
	if resources.MemorySwap != 0 {
		cResources.MemorySwap = resources.MemorySwap
	}
	if resources.MemoryReservation != 0 {
		cResources.MemoryReservation = resources.MemoryReservation
	}
	if resources.KernelMemory != 0 {
		cResources.KernelMemory = resources.KernelMemory
	}

	return nil
}

// Top lists the processes running inside of the given container
func (mgr *ContainerManager) Top(ctx context.Context, name string, psArgs string) (*types.ContainerProcessList, error) {
	if psArgs == "" {
		psArgs = "-ef"
	}

	c, err := mgr.container(name)
	if err != nil {
		return nil, err
	}

	c.Lock()
	defer c.Unlock()

	if !c.IsRunningOrPaused() {
		return nil, fmt.Errorf("container %s is not running or paused, cannot execute top command", c.ID)
	}

	pids, err := mgr.Client.ContainerPIDs(ctx, c.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get pids of container %s", c.ID)
	}

	output, err := exec.Command("ps", strings.Split(psArgs, " ")...).Output()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to run ps command")
	}

	procList, err := parsePSOutput(output, pids)
	if err != nil {
		return nil, errors.Wrapf(err, "failed parsePSOutput")
	}

	return procList, nil
}

// Resize resizes the size of a container tty.
func (mgr *ContainerManager) Resize(ctx context.Context, name string, opts types.ResizeOptions) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if !c.IsRunningOrPaused() {
		return fmt.Errorf("failed to resize container %s: container is not running", c.ID)
	}

	return mgr.Client.ResizeContainer(ctx, c.ID, opts)
}

// Wait stops processing until the given container is stopped.
func (mgr *ContainerManager) Wait(ctx context.Context, name string) (types.ContainerWaitOKBody, error) {
	c, err := mgr.container(name)
	if err != nil {
		return types.ContainerWaitOKBody{}, err
	}

	// We should notice that container's meta data shouldn't be locked in wait process, otherwise waiting for
	// a running container to stop would make other client commands which manage this container are blocked.
	// If a container status is exited or stopped, return exit code immediately.
	if !c.IsRunningOrPaused() {
		return types.ContainerWaitOKBody{
			Error:      c.State.Error,
			StatusCode: c.ExitCode(),
		}, nil
	}

	return mgr.Client.WaitContainer(ctx, c.ID)
}

// Connect is used to connect a container to a network.
func (mgr *ContainerManager) Connect(ctx context.Context, name string, networkIDOrName string, epConfig *types.EndpointSettings) error {
	c, err := mgr.container(name)
	if err != nil {
		return errors.Wrapf(err, "failed to get container %s", name)
	}

	n, err := mgr.NetworkMgr.Get(context.Background(), networkIDOrName)
	if err != nil {
		return errors.Wrapf(err, "failed to get network %s", networkIDOrName)
	}
	if n == nil {
		return fmt.Errorf("network %s does not exist", networkIDOrName)
	}

	if epConfig == nil {
		epConfig = &types.EndpointSettings{}
	}

	c.Lock()
	defer c.Unlock()

	if !c.State.Running {
		if c.State.Dead {
			return fmt.Errorf("container %s is marked for removal and cannot be connected or disconnected to the network %s", c.ID, n.Name)
		}

		if err := mgr.updateNetworkConfig(c, n.Name, epConfig); err != nil {
			return err
		}
	} else if err := mgr.connectToNetwork(ctx, c, networkIDOrName, epConfig); err != nil {
		return err
	}

	mgr.LogNetworkEventWithAttributes(ctx, n.Network, "connect", map[string]string{"container": c.ID})

	return c.Write(mgr.Store)
}

// Disconnect disconnects the given container from
// given network
func (mgr *ContainerManager) Disconnect(ctx context.Context, containerName, networkName string, force bool) error {
	c, err := mgr.container(containerName)
	if err != nil {
		// TODO(ziren): if force is true, force delete endpoint
		return err
	}

	// Get network
	network, err := mgr.NetworkMgr.Get(ctx, networkName)
	if err != nil {
		return fmt.Errorf("failed to get network %s when disconnecting container %s: %v", networkName, c.Name, err)
	}

	c.Lock()
	defer c.Unlock()

	// container cannot be disconnected from host network
	networkMode := c.HostConfig.NetworkMode
	if IsHost(networkMode) && IsHost(network.Mode) {
		return fmt.Errorf("container cannot be disconnected from host network or connected to hostnetwork ")
	}

	if c.NetworkSettings == nil {
		return nil
	}

	epConfig, ok := c.NetworkSettings.Networks[network.Name]
	if !ok {
		// container not attached to the given network
		return fmt.Errorf("failed to disconnect container from network: container %s not attach to %s", c.Name, networkName)
	}

	endpoint := mgr.buildContainerEndpoint(c, network.Name)
	endpoint.EndpointConfig = epConfig
	if err := mgr.NetworkMgr.EndpointRemove(ctx, endpoint); err != nil {
		// TODO(ziren): it is a trick, we should wrapper sandbox
		// not found as an error type
		if !strings.Contains(err.Error(), "not found") {
			logrus.Errorf("failed to remove endpoint: %v", err)
			return err
		}
	}

	// disconnect an endpoint success, delete endpoint info from container json
	delete(c.NetworkSettings.Networks, network.Name)

	// if container has no network attached any more, set NetworkDisabled to true
	// so that not setup Network Namespace when restart the container
	if len(c.NetworkSettings.Networks) == 0 {
		c.Config.NetworkDisabled = true
	}

	// container meta changed, refresh the cache
	mgr.cache.Put(c.ID, c)

	mgr.LogNetworkEventWithAttributes(ctx, network.Network, "disconnect", map[string]string{"container": c.ID})

	// update container meta json
	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update container %s in meta store: %v", c.ID, err)
		return err
	}

	return nil
}

func (mgr *ContainerManager) updateNetworkConfig(container *Container, networkIDOrName string, endpointConfig *types.EndpointSettings) error {
	if IsContainer(container.HostConfig.NetworkMode) {
		return fmt.Errorf("container sharing network namespace with another container or host cannot be connected to any other network")
	}

	// TODO check bridge-mode conflict

	if IsUserDefined(container.HostConfig.NetworkMode) {
		if hasUserDefinedIPAddress(endpointConfig) {
			return fmt.Errorf("user specified IP address is supported on user defined networks only")
		}
		if endpointConfig != nil && len(endpointConfig.Aliases) > 0 {
			return fmt.Errorf("network-scoped alias is supported only for containers in user defined networks")
		}
	} else {
		addShortID := true
		shortID := utils.TruncateID(container.ID)
		for _, alias := range endpointConfig.Aliases {
			if alias == shortID {
				addShortID = false
				break
			}
		}
		if addShortID {
			endpointConfig.Aliases = append(endpointConfig.Aliases, shortID)
		}
	}

	network, err := mgr.NetworkMgr.Get(context.Background(), networkIDOrName)
	if err != nil {
		return err
	}

	if err := validateNetworkingConfig(network.Network, endpointConfig); err != nil {
		return err
	}

	container.NetworkSettings.Networks[network.Name] = endpointConfig

	return nil
}

func (mgr *ContainerManager) connectToNetwork(ctx context.Context, container *Container, networkIDOrName string, epConfig *types.EndpointSettings) (err error) {
	if IsContainer(container.HostConfig.NetworkMode) {
		return fmt.Errorf("container sharing network namespace with another container or host cannot be connected to any other network")
	}

	// TODO check bridge mode conflict

	network, err := mgr.NetworkMgr.Get(context.Background(), networkIDOrName)
	if err != nil {
		return errors.Wrap(err, "failed to get network")
	}

	endpoint := mgr.buildContainerEndpoint(container, network.Name)
	endpoint.EndpointConfig = epConfig
	if _, err := mgr.NetworkMgr.EndpointCreate(ctx, endpoint); err != nil {
		logrus.Errorf("failed to create endpoint: %v", err)
		return err
	}

	return mgr.updateNetworkConfig(container, networkIDOrName, endpoint.EndpointConfig)
}

func (mgr *ContainerManager) initContainerIO(c *Container) (*containerio.IO, error) {
	if io := mgr.IOs.Get(c.ID); io != nil {
		return nil, errors.Wrap(errtypes.ErrConflict, "failed to create containerIO")
	}

	cntrio := containerio.NewIO(c.ID, c.Config.OpenStdin)
	mgr.IOs.Put(c.ID, cntrio)
	return cntrio, nil
}

func (mgr *ContainerManager) initLogDriverBeforeStart(c *Container) error {
	var (
		cntrio *containerio.IO
		err    error
	)

	if cntrio = mgr.IOs.Get(c.ID); cntrio == nil {
		cntrio, err = mgr.initContainerIO(c)
		if err != nil {
			return err
		}
	}

	logInfo, err := mgr.convContainerToLoggerInfo(c)
	if err != nil {
		return err
	}

	logDriver, err := logOptionsForContainerio(c, logInfo)
	if err != nil {
		return err
	}

	if logger.LogMode(logInfo.LogConfig["mode"]) == logger.LogModeNonBlock {
		if maxBufferSize, ok := logInfo.LogConfig["max-buffer-size"]; ok {
			maxBytes, err := units.RAMInBytes(maxBufferSize)
			if err != nil {
				return errors.Wrapf(err, "failed to parse option max-buffer-size: %s", maxBufferSize)
			}
			cntrio.SetMaxBufferSize(maxBytes)
			cntrio.SetNonBlock(true)
		}
	}
	cntrio.SetLogDriver(logDriver)
	return nil
}

func (mgr *ContainerManager) attachCRILog(c *Container, logPath string) error {
	cntrio := mgr.IOs.Get(c.ID)
	if cntrio == nil {
		return errors.Wrap(errtypes.ErrNotfound, "failed to get containerIO")
	}

	return cntrio.AttachCRILog(logPath, c.Config.Tty)
}

func (mgr *ContainerManager) initExecIO(id string, withStdin bool) (*containerio.IO, error) {
	if io := mgr.IOs.Get(id); io != nil {
		return nil, errors.Wrap(errtypes.ErrConflict, "failed to create containerIO")
	}

	cntrio := containerio.NewIO(id, withStdin)
	mgr.IOs.Put(id, cntrio)
	return cntrio, nil
}

func (mgr *ContainerManager) markStoppedAndRelease(c *Container, m *ctrd.Message) error {
	var (
		code   int64  // container exit code used for container state setting
		errMsg string // container exit error message used for container state setting
	)
	if m != nil {
		code = int64(m.ExitCode())
		if err := m.RawError(); err != nil {
			errMsg = err.Error()
		}
	}

	c.SetStatusStopped(code, errMsg)

	// Action Container Remove and function markStoppedAndRelease are conflict.
	// If a container has been removed and the corresponding meta.json will be removed as well.
	// However, when this function markStoppedAndRelease still keeps the container instance,
	// there will be possibility that in markStoppedAndRelease, code calls c.Write(mgr.Store) to
	// write the removed meta.json again. If that, incompatibilty happens.
	// As a result, we check whether this container is still in the meta store.
	if container, err := mgr.container(c.Name); err != nil || container == nil {
		return nil
	}

	// Remove io and network config may occur error, so we should update
	// container's status on disk as soon as possible.
	defer func() {
		if err := c.Write(mgr.Store); err != nil {
			logrus.Errorf("failed to update meta: %v", err)
		}
	}()

	c.UnsetMergedDir()

	return mgr.releaseContainerResources(c)
}

func (mgr *ContainerManager) markExitedAndRelease(c *Container, m *ctrd.Message) error {
	var (
		exitCode int64  // container exit code used for container state setting
		errMsg   string // container exit error message used for container state setting
	)
	if m != nil {
		exitCode = int64(m.ExitCode())
		if err := m.RawError(); err != nil {
			errMsg = err.Error()
		}
	}

	c.SetStatusExited(exitCode, errMsg)

	// Action Container Remove and function markStoppedAndRelease are conflict.
	// If a container has been removed and the corresponding meta.json will be removed as well.
	// However, when this function markStoppedAndRelease still keeps the container instance,
	// there will be possibility that in markStoppedAndRelease, code calls c.Write(mgr.Store) to
	// write the removed meta.json again. If that, incompatibilty happens.
	// As a result, we check whether this container is still in the meta store.
	if container, err := mgr.container(c.Name); err != nil || container == nil {
		return nil
	}

	// Remove io and network config may occur error, so we should update
	// container's status on disk as soon as possible.
	defer func() {
		if err := c.Write(mgr.Store); err != nil {
			logrus.Errorf("failed to update meta: %v", err)
		}
	}()

	c.UnsetMergedDir()

	return mgr.releaseContainerResources(c)
}

// exitedAndRelease be register into ctrd as a callback function, when the running container suddenly
// exited, "ctrd" will call it to set the container's state and release resouce and so on.
func (mgr *ContainerManager) exitedAndRelease(id string, m *ctrd.Message, cleanup func() error) error {
	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if err := mgr.markExitedAndRelease(c, m); err != nil {
		return err
	}

	// for example, delete containerd container/task.
	if cleanup != nil {
		if err := cleanup(); err != nil {
			return err
		}
	}

	// Action Container Remove and function exitedAndRelease are conflict.
	// If a container has been removed and the corresponding meta.json will be removed as well.
	// However, when this function exitedAndRelease still keeps the container instance,
	// there will be possibility that in exitedAndRelease, code calls c.Write(mgr.Store) to
	// write the removed meta.json again. If that, incompatibilty happens.
	// As a result, we check whether this container is still in the meta store.
	if container, err := mgr.container(c.Name); err != nil || container == nil {
		return nil
	}

	// send exit event to monitor
	mgr.monitor.PostEvent(ContainerExitEvent(c).WithHandle(func(c *Container) error {
		// check status and restart policy
		if !c.State.Exited {
			return nil
		}

		policy := (*ContainerRestartPolicy)(c.HostConfig.RestartPolicy)
		keys := c.DetachKeys

		if policy == nil || policy.IsNone() {
			return nil
		}

		return mgr.Start(context.TODO(), c.ID, &types.ContainerStartOptions{DetachKeys: keys})
	}))

	return nil
}

// execExitedAndRelease be register into ctrd as a callback function, when the exec process in a container
// exited, "ctrd" will call it to release resource and so on.
func (mgr *ContainerManager) execExitedAndRelease(id string, m *ctrd.Message) error {
	v, ok := mgr.ExecProcesses.Get(id).Result()
	if !ok {
		return errors.Wrapf(errtypes.ErrNotfound, "exec process %s", id)
	}
	execConfig, ok := v.(*ContainerExecConfig)
	if !ok {
		return fmt.Errorf("invalid exec config type")
	}

	execConfig.ExitCode = int64(m.ExitCode())
	execConfig.Running = false
	execConfig.Error = m.RawError()

	eio := mgr.IOs.Get(id)
	if eio == nil {
		return nil
	}

	// close io
	eio.Close()
	mgr.IOs.Remove(id)
	return nil
}

func (mgr *ContainerManager) releaseContainerResources(c *Container) error {
	mgr.resetContainerIOs(c.ID)
	return mgr.releaseContainerNetwork(c)
}

// releaseContainerNetwork release container network when container exits or is stopped.
func (mgr *ContainerManager) releaseContainerNetwork(c *Container) error {
	// NetworkMgr is nil, which means the pouch daemon is initializing.
	// And the libnetwork will also initialize, which will release all
	// staled network resources(endpoint, network and namespace). So we
	// don't need release the network resources.
	if mgr.NetworkMgr == nil || c.NetworkSettings == nil {
		return nil
	}

	for name, epConfig := range c.NetworkSettings.Networks {
		endpoint := mgr.buildContainerEndpoint(c, name)
		endpoint.EndpointConfig = epConfig
		if err := mgr.NetworkMgr.EndpointRemove(context.Background(), endpoint); err != nil {
			// TODO(ziren): it is a trick, we should wrapper "sanbox
			// not found"" as an error type
			if !strings.Contains(err.Error(), "not found") {
				logrus.Errorf("failed to remove endpoint: %v", err)
				return err
			}
		}
	}

	return nil
}

// resetContainerIOs resets container IO resources.
func (mgr *ContainerManager) resetContainerIOs(containerID string) {
	// release resource
	io := mgr.IOs.Get(containerID)
	if io == nil {
		return
	}

	io.Reset()
}

// buildContainerEndpoint builds Endpoints according to container
// caller should lock container when calling this func.
func (mgr *ContainerManager) buildContainerEndpoint(c *Container, name string) *networktypes.Endpoint {
	ep := BuildContainerEndpoint(c)
	ep.Name = name

	if !IsUserDefined(name) {
		ep.DisableResolver = true
	}

	if mgr.containerPlugin != nil {
		// just ignore return err
		err := mgr.containerPlugin.PreCreateEndpoint(c.ID, c.Config.Env, ep)
		if err != nil {
			logrus.Warnf("failed to call PreCreateEndpoint plugin, err(%v)", err)
		}
	}

	return ep
}

// setBaseFS keeps container basefs in meta.
func (mgr *ContainerManager) setBaseFS(ctx context.Context, c *Container) {
	snapshotID := c.SnapshotKey()
	_, err := mgr.Client.GetSnapshot(ctx, snapshotID)
	if err != nil {
		logrus.Errorf("failed to get container %s snapshot %s: %v", c.Key(), snapshotID, err)
		return
	}

	var managerID string
	if c.HostConfig.RuntimeType == ctrd.RuntimeTypeV1 {
		managerID = ctrd.RuntimeTypeV1
	} else {
		managerID = "io.containerd.runtime.v2.task"
	}

	c.BaseFS = filepath.Join(mgr.Config.HomeDir, "containerd/state", managerID, mgr.Config.DefaultNamespace, c.ID, "rootfs")
}

// execProcessGC cleans unused exec processes config every 5 minutes.
func (mgr *ContainerManager) execProcessGC() {
	for range time.Tick(time.Duration(GCExecProcessTick) * time.Minute) {
		execProcesses := mgr.ExecProcesses.Values(nil)
		cleaned := 0

		for id, v := range execProcesses {
			execConfig, ok := v.(*ContainerExecConfig)
			if !ok {
				logrus.Warnf("get incorrect exec config: %v", v)
				continue
			}
			// if unused exec processes are found, we will tag them, and clean
			// them in next loop, so that we can ensure exec process can get
			// correct exit code.
			if execConfig.WaitForClean {
				cleaned++
				mgr.ExecProcesses.Remove(id)
			} else if !execConfig.Running {
				execConfig.WaitForClean = true
			}
		}

		if cleaned > 0 {
			logrus.Debugf("clean %d unused exec process", cleaned)
		}
	}
}

// NewSnapshotsSyncer creates a snapshot syncer.
func (mgr *ContainerManager) NewSnapshotsSyncer(snapshotStore *SnapshotStore, duration time.Duration) *SnapshotsSyncer {
	return newSnapshotsSyncer(snapshotStore, mgr.Client, duration)
}

func (mgr *ContainerManager) generateContainerID(specificID string) (string, error) {
	if specificID != "" {
		if len(specificID) != 64 {
			return "", errors.Wrap(errtypes.ErrInvalidParam, "Container id length should be 64")
		}
		//characters of containerID should be in "0123456789abcdef"
		for _, c := range []byte(specificID) {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				return "", errors.Wrap(errtypes.ErrInvalidParam, "The characters of container id should be in '0123456789abcdef'")
			}
		}

		if mgr.cache.Get(specificID).Exist() {
			return "", errors.Wrap(errtypes.ErrAlreadyExisted, "container id: "+specificID)
		}
		return specificID, nil
	}

	return mgr.generateID()
}
