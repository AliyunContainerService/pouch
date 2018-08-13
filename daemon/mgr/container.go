package mgr

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/opts"
	"github.com/alibaba/pouch/apis/plugins"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/daemon/events"
	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/lxcfs"
	networktypes "github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	mountutils "github.com/alibaba/pouch/pkg/mount"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/storage/quota"
	volumetypes "github.com/alibaba/pouch/storage/volume/types"

	containerdtypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/mount"
	"github.com/docker/libnetwork"
	"github.com/go-openapi/strfmt"
	"github.com/imdario/mergo"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ContainerMgr as an interface defines all operations against container.
// ContainerMgr's functionality could be divided into three parts:
// 1. regular container management;
// 2. container exec management;
// 3. container network management.
type ContainerMgr interface {
	// 1. the following functions are related to regular container management

	// Restore containers from meta store to memory and recover those container.
	Restore(ctx context.Context) error

	// Create a new container.
	Create(ctx context.Context, name string, config *types.ContainerCreateConfig) (*types.ContainerCreateResp, error)

	// Get the detailed information of container.
	Get(ctx context.Context, name string) (*Container, error)

	// List returns the list of containers.
	List(ctx context.Context, option *ContainerListOption) ([]*Container, error)

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

	// Stats of a container.
	Stats(ctx context.Context, name string) (*containerdtypes.Metric, error)

	// Attach a container.
	Attach(ctx context.Context, name string, attach *AttachConfig) error

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
	StartExec(ctx context.Context, execid string, attach *AttachConfig) error

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

	containerPlugin plugins.ContainerPlugin

	// eventsService is used to publish events generated by pouchd
	eventsService *events.Events
}

// NewContainerManager creates a brand new container manager.
func NewContainerManager(ctx context.Context, store *meta.Store, cli ctrd.APIClient, imgMgr ImageMgr, volMgr VolumeMgr, cfg *config.Config, contPlugin plugins.ContainerPlugin, eventsService *events.Events) (*ContainerManager, error) {
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
	mgr.Client.SetEventsHooks(mgr.publishContainerdEvent)

	go mgr.execProcessGC()

	return mgr, nil
}

// Restore containers from meta store to memory and recover those container.
func (mgr *ContainerManager) Restore(ctx context.Context) error {
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

		if container.State.Status != types.StatusRunning &&
			container.State.Status != types.StatusPaused {
			return nil
		}

		// recover the running or paused container.
		io, err := mgr.openContainerIO(container)
		if err != nil {
			logrus.Errorf("failed to recover container: %s,  %v", id, err)
		}

		err = mgr.Client.RecoverContainer(ctx, id, io)
		if err != nil && strings.Contains(err.Error(), "not found") {
			logrus.Infof("container %s not found, executes mark stopped and release resources", id)
			if err := mgr.markStoppedAndRelease(container, nil); err != nil {
				logrus.Errorf("failed to mark container %s stop status: %v", id, err)
			}
		} else if err != nil {
			logrus.Errorf("failed to recover container: %s,  %v", id, err)
			// release io
			io.Close()
			mgr.IOs.Remove(id)
		}

		return nil
	}
	return mgr.Store.ForEach(fn)
}

// Create checks passed in parameters and create a Container object whose status is set at Created.
func (mgr *ContainerManager) Create(ctx context.Context, name string, config *types.ContainerCreateConfig) (resp *types.ContainerCreateResp, err error) {
	// cleanup allocated resources when failed
	cleanups := []func(){}
	defer func() {
		// do cleanup
		if err != nil {
			for _, f := range cleanups {
				f()
			}
		}
	}()

	imgID, _, primaryRef, err := mgr.ImageMgr.CheckReference(ctx, config.Image)
	if err != nil {
		return nil, err
	}

	// TODO: check request validate.
	if config.HostConfig == nil {
		return nil, fmt.Errorf("HostConfig cannot be nil")
	}
	if config.NetworkingConfig == nil {
		return nil, fmt.Errorf("NetworkingConfig cannot be nil")
	}

	id, err := mgr.generateID()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = mgr.generateName(id)
	} else if mgr.NameToID.Get(name).Exist() {
		return nil, errors.Wrap(errtypes.ErrAlreadyExisted, "container name: "+name)
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

	if _, exist := mgr.Config.Runtimes[config.HostConfig.Runtime]; !exist {
		return nil, fmt.Errorf("unknown runtime %s", config.HostConfig.Runtime)
	}

	config.Image = primaryRef.String()
	// create a snapshot with image.
	if err := mgr.Client.CreateSnapshot(ctx, id, config.Image); err != nil {
		return nil, err
	}
	cleanups = append(cleanups, func() {
		mgr.Client.RemoveSnapshot(ctx, id)
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
	}

	// set container basefs, basefs is not created in pouchd, it will created
	// after create options passed to containerd.
	mgr.setBaseFS(ctx, container, id)

	if err := mgr.Mount(ctx, container); err != nil {
		return nil, err
	}

	// parse volume config
	if err := mgr.generateMountPoints(ctx, container); err != nil {
		return nil, errors.Wrap(err, "failed to parse volume argument")
	}

	// set mount point disk quota
	if err := mgr.setMountPointDiskQuota(ctx, container); err != nil {
		return nil, errors.Wrap(err, "failed to set mount point disk quota")
	}

	// set rootfs disk quota
	if err := mgr.setRootfsQuota(ctx, container); err != nil {
		logrus.Warnf("failed to set rootfs disk quota, err: %v", err)
	}

	if err := mgr.Unmount(ctx, container); err != nil {
		return nil, err
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
	if container.NetworkSettings.Networks == nil && !IsContainer(config.HostConfig.NetworkMode) {
		container.NetworkSettings.Networks = make(map[string]*types.EndpointSettings)
		container.NetworkSettings.Networks[config.HostConfig.NetworkMode] = new(types.EndpointSettings)
	}

	if err := parseSecurityOpts(container, config.HostConfig.SecurityOpt); err != nil {
		return nil, err
	}

	// merge image's config into container
	if err := container.merge(func() (ocispec.ImageConfig, error) {
		img, err := mgr.Client.GetImage(ctx, config.Image)
		ociImage, err := containerdImageToOciImage(ctx, img)
		if err != nil {
			return ocispec.ImageConfig{}, err
		}
		return ociImage.Config, nil
	}); err != nil {
		return nil, err
	}

	// Get snapshot UpperDir
	mounts, err := mgr.Client.GetMounts(ctx, id)
	if err != nil {
		return nil, err
	} else if len(mounts) != 1 {
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

	return c, nil
}

// Start a pre created Container.
func (mgr *ContainerManager) Start(ctx context.Context, id string, options *types.ContainerStartOptions) (err error) {
	if id == "" {
		return errors.Wrap(errtypes.ErrInvalidParam, "either container name or id is required")
	}

	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	return mgr.start(ctx, c, options)
}

func (mgr *ContainerManager) start(ctx context.Context, c *Container, options *types.ContainerStartOptions) error {
	var err error
	c.DetachKeys = options.DetachKeys

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

	mgr.LogContainerEvent(ctx, c, "start")

	return nil
}

func (mgr *ContainerManager) prepareContainerNetwork(ctx context.Context, c *Container) error {
	c.Lock()
	defer c.Unlock()

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

	// initialise network endpoint
	if c.NetworkSettings == nil {
		return nil
	}

	for name, endpointSetting := range c.NetworkSettings.Networks {
		endpoint := mgr.buildContainerEndpoint(c)
		endpoint.Name = name
		endpoint.EndpointConfig = endpointSetting
		if _, err := mgr.NetworkMgr.EndpointCreate(ctx, endpoint); err != nil {
			logrus.Errorf("failed to create endpoint: %v", err)
			return err
		}
	}

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
	c.Lock()
	if c.HostConfig.CgroupParent == "" {
		c.HostConfig.CgroupParent = mgr.Config.CgroupParent
	}
	c.Unlock()

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
		ctrMgr:  mgr,
		volMgr:  mgr.VolumeMgr,
		netMgr:  mgr.NetworkMgr,
		prioArr: prioArr,
		argsArr: argsArr,
	}

	if err = createSpec(ctx, c, sw); err != nil {
		return err
	}

	// open container's stdio.
	io, err := mgr.openContainerIO(c)
	if err != nil {
		return errors.Wrap(err, "failed to open io")
	}

	c.Lock()
	ctrdContainer := &ctrd.Container{
		ID:             c.ID,
		Image:          c.Config.Image,
		Labels:         c.Config.Labels,
		Runtime:        c.HostConfig.Runtime,
		Spec:           sw.s,
		IO:             io,
		RootFSProvided: c.RootFSProvided,
		BaseFS:         c.BaseFS,
	}
	c.Unlock()

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
	c.Lock()
	c.Snapshotter.Data["MergedDir"] = c.BaseFS
	c.Unlock()

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

	if !c.IsRunningOrPaused() {
		// stopping a non-running container is valid.
		return nil
	}

	if timeout == 0 {
		timeout = c.StopTimeout()
	}

	err = mgr.stop(ctx, c, timeout)
	if err != nil {
		return err
	}
	mgr.LogContainerEvent(ctx, c, "stop")

	return nil
}

func (mgr *ContainerManager) stop(ctx context.Context, c *Container, timeout int64) error {
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

	if timeout == 0 {
		timeout = c.StopTimeout()
	}

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
	return mgr.start(ctx, c, &types.ContainerStartOptions{})
}

// Pause pauses a running container.
func (mgr *ContainerManager) Pause(ctx context.Context, name string) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	if !c.IsRunning() {
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

	if !c.IsPaused() {
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

// Stats gets the stat of a container.
func (mgr *ContainerManager) Stats(ctx context.Context, name string) (*containerdtypes.Metric, error) {
	var (
		err error
		c   *Container
	)

	if c, err = mgr.container(name); err != nil {
		return nil, err
	}

	c.Lock()
	defer c.Unlock()

	return mgr.Client.ContainerStats(ctx, c.ID)
}

// Attach attachs a container's io.
func (mgr *ContainerManager) Attach(ctx context.Context, name string, attach *AttachConfig) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	_, err = mgr.openAttachIO(c, attach)
	if err != nil {
		return err
	}

	return nil
}

// Rename renames a container.
func (mgr *ContainerManager) Rename(ctx context.Context, oldName, newName string) error {
	if mgr.NameToID.Get(newName).Exist() {
		return errors.Wrap(errtypes.ErrAlreadyExisted, "container name: "+newName)
	}

	c, err := mgr.container(oldName)
	if err != nil {
		return errors.Wrapf(err, "failed to rename container %s", oldName)
	}

	attributes := map[string]string{
		"oldName": oldName,
	}

	c.Lock()
	name := c.Name
	c.Name = newName
	c.Unlock()

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

	warnings, err := validateResource(&config.Resources, true)
	if err != nil {
		return err
	}
	if len(warnings) != 0 {
		logrus.Warnf("warnings update %s: %v", name, warnings)
	}

	restore := false
	configBack := *c.Config
	hostconfigBack := *c.HostConfig
	defer func() {
		if restore {
			c.Lock()
			c.Config = &configBack
			c.HostConfig = &hostconfigBack
			c.Unlock()
		}
	}()

	if c.IsRunning() && config.Resources.KernelMemory != 0 {
		return fmt.Errorf("failed to update container %s: can not update kernel memory to a running container, please stop it first", c.ID)
	}

	// update container disk quota
	if err := mgr.updateContainerDiskQuota(ctx, c, config.DiskQuota); err != nil {
		return errors.Wrapf(err, "failed to update diskquota of container %s", c.ID)
	}

	c.Lock()

	// init Container Labels
	if c.Config.Labels == nil {
		c.Config.Labels = map[string]string{}
	}

	// compatibility with alidocker, UpdateConfig.Label is []string
	// but ContainerConfig.Labels is map[string]string
	if len(config.Label) != 0 {
		// support remove some labels
		newLabels, err := opts.ParseLabels(config.Label)
		if err != nil {
			return errors.Wrapf(err, "failed to parse labels")
		}

		for k, v := range newLabels {
			if v == "" {
				delete(c.Config.Labels, k)
			} else {
				c.Config.Labels[k] = v
			}
		}
	}

	// TODO(ziren): we should use meta.Config.DiskQuota to record container diskquota
	// compatibility with alidocker, when set DiskQuota for container
	// add a DiskQuota label
	if config.DiskQuota != nil {
		if _, ok := c.Config.Labels["DiskQuota"]; ok {
			labels := []string{}
			for dir, quota := range c.Config.DiskQuota {
				labels = append(labels, fmt.Sprintf("%s=%s", dir, quota))
			}
			c.Config.Labels["DiskQuota"] = strings.Join(labels, ";")
		}
	}

	c.Unlock()

	// update Resources of a container.
	if err := mgr.updateContainerResources(c, config.Resources); err != nil {
		restore = true
		return errors.Wrapf(err, "failed to update resource of container %s", c.ID)
	}

	c.Lock()
	// TODO update restartpolicy when container is running.
	if config.RestartPolicy != nil && config.RestartPolicy.Name != "" {
		c.HostConfig.RestartPolicy = config.RestartPolicy
	}
	c.Unlock()

	// update env when container is running, default snapshotter driver
	// is overlayfs
	if c.IsRunningOrPaused() && len(config.Env) > 0 && c.Snapshotter != nil {
		if mergedDir, exists := c.Snapshotter.Data["MergedDir"]; exists {
			if err := updateContainerEnv(c.Config.Env, mergedDir); err != nil {
				return errors.Wrapf(err, "failed to update env of running container")
			}
		}
	}

	// Update Env
	if len(config.Env) > 0 {
		newEnvMap, err := opts.ParseEnv(config.Env)
		if err != nil {
			return errors.Wrapf(err, "failed to parse new env")
		}

		oldEnvMap, err := opts.ParseEnv(c.Config.Env)
		if err != nil {
			return errors.Wrapf(err, "failed to parse old env")
		}

		for k, v := range newEnvMap {
			// key should not be empty
			if k == "" {
				continue
			}

			// add or change an env
			if v != "" {
				oldEnvMap[k] = v
				continue
			}

			// value is empty, we need delete the env
			if _, exists := oldEnvMap[k]; exists {
				delete(oldEnvMap, k)
			}
		}

		newEnvSlice := []string{}
		for k, v := range oldEnvMap {
			newEnvSlice = append(newEnvSlice, fmt.Sprintf("%s=%s", k, v))
		}

		c.Config.Env = newEnvSlice
	}

	// If container is not running, update container metadata struct is enough,
	// resources will be updated when the container is started again,
	// If container is running, we need to update configs to the real world.
	if c.IsRunning() {
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

	if !c.IsStopped() && !c.IsExited() && !c.IsCreated() && !options.Force {
		return fmt.Errorf("container %s is not stopped, cannot remove it without flag force", c.ID)
	}

	// if the container is running, force to stop it.
	if c.IsRunning() && options.Force {
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
	// there is no snapshot for this container.
	if !c.RootFSProvided {
		// remove snapshot
		if err := mgr.Client.RemoveSnapshot(ctx, c.ID); err != nil {
			logrus.Errorf("failed to remove snapshot of container %s: %v", c.ID, err)
		}
	}

	// When removing a container, we have set up such rule for object removing sequences:
	// 1. container object in pouchd's memory;
	// 2. meta.json for container in local disk.

	// remove name
	mgr.NameToID.Remove(c.Name)
	// remove container cache
	mgr.cache.Remove(c.ID)

	// remove meta.json for container in local disk
	if err := mgr.Store.Remove(c.Key()); err != nil {
		logrus.Errorf("failed to remove container %s from meta store: %v", c.ID, err)
	}

	mgr.LogContainerEvent(ctx, c, "destroy")
	return nil
}

func (mgr *ContainerManager) updateContainerDiskQuota(ctx context.Context, c *Container, diskQuota map[string]string) error {
	if diskQuota == nil {
		return nil
	}

	c.Lock()
	for dir, quota := range diskQuota {
		c.Config.DiskQuota[dir] = quota
	}
	c.Unlock()

	// set mount point disk quota
	if err := mgr.setMountPointDiskQuota(ctx, c); err != nil {
		return errors.Wrapf(err, "failed to set mount point disk quota")
	}

	c.Lock()
	var qid uint32
	if c.Config.QuotaID != "" {
		id, err := strconv.Atoi(c.Config.QuotaID)
		if err != nil {
			return errors.Wrapf(err, "failed to convert QuotaID %s", c.Config.QuotaID)
		}

		qid = uint32(id)
		if id < 0 {
			// QuotaID is < 0, it means pouchd alloc a unique quota id.
			qid, err = quota.GetNextQuotaID()
			if err != nil {
				return errors.Wrap(err, "failed to get next quota id")
			}

			// update QuotaID
			c.Config.QuotaID = strconv.Itoa(int(qid))
		}
	}
	c.Unlock()

	// get rootfs quota
	defaultQuota := quota.GetDefaultQuota(c.Config.DiskQuota)
	if qid > 0 && defaultQuota == "" {
		return fmt.Errorf("set quota id but have no set default quota size")
	}
	// update container rootfs disk quota
	// TODO: add lock for container?
	if c.IsRunningOrPaused() && c.Snapshotter != nil {
		basefs, ok := c.Snapshotter.Data["MergedDir"]
		if !ok || basefs == "" {
			return fmt.Errorf("Container is running, but MergedDir is missing")
		}

		if err := quota.SetRootfsDiskQuota(basefs, defaultQuota, qid); err != nil {
			return errors.Wrapf(err, "failed to set container rootfs diskquota")
		}
	}

	return nil
}

// updateContainerResources update container's resources parameters.
func (mgr *ContainerManager) updateContainerResources(c *Container, resources types.Resources) error {
	c.Lock()
	defer c.Unlock()
	// update resources of container.
	cResources := &c.HostConfig.Resources
	if resources.BlkioWeight != 0 {
		cResources.BlkioWeight = resources.BlkioWeight
	}
	if resources.CPUPeriod != 0 {
		cResources.CPUPeriod = resources.CPUPeriod
	}
	if resources.CPUQuota > -1 {
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

// Upgrade upgrades a container with new image and args.
func (mgr *ContainerManager) Upgrade(ctx context.Context, name string, config *types.ContainerUpgradeConfig) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	// check the image existed or not, and convert image id to image ref
	_, _, primaryRef, err := mgr.ImageMgr.CheckReference(ctx, config.Image)
	if err != nil {
		return errors.Wrap(err, "failed to get image")
	}
	config.Image = primaryRef.String()

	// Nothing changed, no need upgrade.
	if config.Image == c.Config.Image {
		return fmt.Errorf("failed to upgrade container: image not changed")
	}

	var (
		needRollback = false
		// FIXME: do a deep copy of container?
		backupContainer = c
	)

	defer func() {
		if needRollback {
			c = backupContainer
			if err := mgr.Client.CreateSnapshot(ctx, c.ID, c.Image); err != nil {
				logrus.Errorf("failed to create snapshot when rollback upgrade action: %v", err)
				return
			}
			// FIXME: create new containerd container may failed
			_ = mgr.createContainerdContainer(ctx, c, "", "")
		}
	}()

	// FIXME(ziren): mergo.Merge() use AppendSlice to merge slice.
	// that is to say, t1 = ["a", "b"], t2 = ["a", "c"], the merge
	// result will be ["a", "b", "a", "c"]
	// This may occur errors, just take notes to record this.
	if err := mergo.MergeWithOverwrite(c.Config, config.ContainerConfig); err != nil {
		return errors.Wrapf(err, "failed to merge ContainerConfig")
	}
	if err := mergo.MergeWithOverwrite(c.HostConfig, config.HostConfig); err != nil {
		return errors.Wrapf(err, "failed to merge HostConfig")
	}
	c.Image = config.Image

	// If container is not running, we just store this data.
	if c.State.Status != types.StatusRunning {
		// Works fine, store new container info to disk.
		if err := c.Write(mgr.Store); err != nil {
			logrus.Errorf("failed to update container %s in meta store: %v", c.ID, err)
			return err
		}
		return nil
	}
	// If container is running,  we need change
	// configuration and recreate it. Else we just store new meta
	// into disk, next time when starts container, the new configurations
	// will take effect.

	// Inherit volume configurations from old container,
	// New volume configurations may cover the old one.
	// c.VolumesFrom = []string{c.ID}

	// FIXME(ziren): here will forcely stop container afer 3s.
	// If DestroyContainer failed, we think the old container
	// not changed, so just return error, no need recover it.
	if _, err := mgr.Client.DestroyContainer(ctx, c.ID, 3); err != nil {
		return errors.Wrapf(err, "failed to destroy container")
	}

	// remove snapshot of old container
	if err := mgr.Client.RemoveSnapshot(ctx, c.ID); err != nil {
		return errors.Wrap(err, "failed to remove snapshot")
	}

	// wait util old snapshot to be deleted
	wait := make(chan struct{})
	go func() {
		for {
			// FIXME(ziren) Ensure the removed snapshot be removed
			// by garbage collection.
			time.Sleep(100 * time.Millisecond)

			_, err := mgr.Client.GetSnapshot(ctx, c.ID)
			if err != nil && errdefs.IsNotFound(err) {
				close(wait)
				return
			}
		}
	}()

	select {
	case <-wait:
		// TODO delete snapshot succeeded
	case <-time.After(30 * time.Second):
		needRollback = true
		return fmt.Errorf("failed to deleted old snapshot: wait old snapshot %s to be deleted timeout(30s)", c.ID)
	}

	// create a snapshot with image for new container.
	if err := mgr.Client.CreateSnapshot(ctx, c.ID, config.Image); err != nil {
		needRollback = true
		return errors.Wrap(err, "failed to create snapshot")
	}

	if err := mgr.createContainerdContainer(ctx, c, "", ""); err != nil {
		needRollback = true
		return errors.Wrap(err, "failed to create new container")
	}

	// Upgrade succeeded, refresh the cache
	mgr.cache.Put(c.ID, c)

	// Works fine, store new container info to disk.
	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update container %s in meta store: %v", c.ID, err)
		return err
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

	if !c.IsRunning() {
		return nil, fmt.Errorf("container %s is not running, cannot execute top command", c.ID)
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
	if c.IsExited() || c.IsStopped() {
		return types.ContainerWaitOKBody{
			Error:      c.State.Error,
			StatusCode: c.ExitCode(),
		}, nil
	}
	// If a container status is created, return 0 as status code.
	if c.IsCreated() {
		return types.ContainerWaitOKBody{
			StatusCode: 0,
		}, nil
	}

	return mgr.Client.WaitContainer(ctx, c.ID)
}

// Connect is used to connect a container to a network.
func (mgr *ContainerManager) Connect(ctx context.Context, name string, networkIDOrName string, epConfig *types.EndpointSettings) error {
	c, err := mgr.container(name)
	if err != nil {
		return errors.Wrapf(err, "failed to get container: %s", name)
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

	if !c.IsRunning() {
		if c.IsDead() {
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

	// container cannot be disconnected from host network
	c.Lock()
	networkMode := c.HostConfig.NetworkMode
	c.Unlock()

	if IsHost(networkMode) && IsHost(network.Mode) {
		return fmt.Errorf("container cannot be disconnected from host network or connected to hostnetwork ")
	}

	c.Lock()
	networkSettings := c.NetworkSettings
	c.Unlock()

	if networkSettings == nil {
		return nil
	}

	epConfig, ok := networkSettings.Networks[network.Name]
	if !ok {
		// container not attached to the given network
		return fmt.Errorf("failed to disconnect container from network: container %s not attach to %s", c.Name, networkName)
	}

	c.Lock()
	endpoint := mgr.buildContainerEndpoint(c)
	c.Unlock()
	endpoint.Name = network.Name
	endpoint.EndpointConfig = epConfig
	if err := mgr.NetworkMgr.EndpointRemove(ctx, endpoint); err != nil {
		// TODO(ziren): it is a trick, we should wrapper sanbox
		// not found as an error type
		if !strings.Contains(err.Error(), "not found") {
			logrus.Errorf("failed to remove endpoint: %v", err)
			return err
		}
	}

	// disconnect an endpoint success, delete endpoint info from container json
	delete(networkSettings.Networks, network.Name)

	// if container has no network attached any more, set NetworkDisabled to true
	// so that not setup Network Namespace when restart the container
	c.Lock()
	if len(networkSettings.Networks) == 0 {
		c.Config.NetworkDisabled = true
	}

	// container meta changed, refresh the cache
	mgr.cache.Put(c.ID, c)
	c.Unlock()

	mgr.LogNetworkEventWithAttributes(ctx, network.Network, "disconnect", map[string]string{"container": c.ID})

	// update container meta json
	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update container %s in meta store: %v", c.ID, err)
		return err
	}

	return nil
}

func (mgr *ContainerManager) openContainerIO(c *Container) (*containerio.IO, error) {
	if io := mgr.IOs.Get(c.ID); io != nil {
		return io, nil
	}

	logInfo := mgr.convContainerToLoggerInfo(c)
	options := []func(*containerio.Option){
		containerio.WithID(c.ID),
		containerio.WithLoggerInfo(logInfo),
		containerio.WithStdin(c.Config.OpenStdin),
	}

	options = append(options, logOptionsForContainerio(c)...)

	io := containerio.NewIO(containerio.NewOption(options...))
	mgr.IOs.Put(c.ID, io)
	return io, nil
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

	container.Lock()
	endpoint := mgr.buildContainerEndpoint(container)
	container.Unlock()
	endpoint.Name = network.Name
	endpoint.EndpointConfig = epConfig
	if _, err := mgr.NetworkMgr.EndpointCreate(ctx, endpoint); err != nil {
		logrus.Errorf("failed to create endpoint: %v", err)
		return err
	}

	return mgr.updateNetworkConfig(container, networkIDOrName, endpoint.EndpointConfig)
}

// FIXME: remove this useless functions
func (mgr *ContainerManager) updateNetworkSettings(container *Container, n libnetwork.Network) error {
	if container.NetworkSettings == nil {
		container.NetworkSettings = &types.NetworkSettings{Networks: make(map[string]*types.EndpointSettings)}
	}

	if !IsHost(container.HostConfig.NetworkMode) && IsHost(n.Type()) {
		return fmt.Errorf("container cannot be connected to host network")
	}

	for s := range container.NetworkSettings.Networks {
		sn, err := mgr.NetworkMgr.Get(context.Background(), s)
		if err != nil {
			continue
		}

		if sn.Name == n.Name() {
			// Avoid duplicate config
			return nil
		}
		if !IsPrivate(sn.Type) || !IsPrivate(n.Type()) {
			return fmt.Errorf("container sharing network namespace with another container or host cannot be connected to any other network")
		}
		if IsNone(sn.Name) || IsNone(n.Name()) {
			return fmt.Errorf("container cannot be connected to multiple networks with one of the networks in none mode")
		}
	}

	if _, ok := container.NetworkSettings.Networks[n.Name()]; !ok {
		container.NetworkSettings.Networks[n.Name()] = new(types.EndpointSettings)
	}

	return nil
}

func (mgr *ContainerManager) openExecIO(id string, attach *AttachConfig) (*containerio.IO, error) {
	if io := mgr.IOs.Get(id); io != nil {
		return io, nil
	}

	options := []func(*containerio.Option){
		containerio.WithID(id),
	}

	if attach != nil {
		options = append(options, attachConfigToOptions(attach)...)
		options = append(options, containerio.WithStdin(attach.Stdin))
		options = append(options, containerio.WithMuxDisabled(attach.MuxDisabled))
	} else {
		options = append(options, containerio.WithDiscard())
	}

	io := containerio.NewIO(containerio.NewOption(options...))
	mgr.IOs.Put(id, io)
	return io, nil
}

func (mgr *ContainerManager) openAttachIO(c *Container, attach *AttachConfig) (*containerio.IO, error) {
	logInfo := mgr.convContainerToLoggerInfo(c)
	options := []func(*containerio.Option){
		containerio.WithID(c.ID),
		containerio.WithLoggerInfo(logInfo),
	}
	options = append(options, logOptionsForContainerio(c)...)

	if attach != nil {
		options = append(options, attachConfigToOptions(attach)...)
		options = append(options, containerio.WithStdin(attach.Stdin))
	} else {
		options = append(options, containerio.WithDiscard())
	}

	io := mgr.IOs.Get(c.ID)
	if io != nil {
		io.AddBackend(containerio.NewOption(options...))
	} else {
		io = containerio.NewIO(containerio.NewOption(options...))
	}
	mgr.IOs.Put(c.ID, io)
	return io, nil
}

func attachConfigToOptions(attach *AttachConfig) []func(*containerio.Option) {
	options := []func(*containerio.Option){}
	if attach.Hijack != nil {
		// Attaching using http.
		options = append(options, containerio.WithHijack(attach.Hijack, attach.Upgrade))
		if attach.Stdin {
			options = append(options, containerio.WithStdinHijack())
		}
	} else if attach.Pipe != nil {
		// Attaching using pipe.
		options = append(options, containerio.WithPipe(attach.Pipe))
	} else if attach.Streams != nil {
		// Attaching using streams.
		options = append(options, containerio.WithStreams(attach.Streams))
		if attach.Stdin {
			options = append(options, containerio.WithStdinStream())
		}
	}

	if attach.CriLogFile != nil {
		options = append(options, containerio.WithCriLogFile(attach.CriLogFile))
	}

	return options
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
func (mgr *ContainerManager) exitedAndRelease(id string, m *ctrd.Message) error {
	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	if err := mgr.markExitedAndRelease(c, m); err != nil {
		return err
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
		if !c.IsExited() {
			return nil
		}

		c.Lock()
		policy := (*ContainerRestartPolicy)(c.HostConfig.RestartPolicy)
		keys := c.DetachKeys
		c.Unlock()

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

	io := mgr.IOs.Get(id)
	if io == nil {
		return nil
	}

	if err := m.RawError(); err != nil {
		fmt.Fprintf(io.Stdout, "%v\n", err)
	}

	// close io
	io.Close()
	mgr.IOs.Remove(id)

	return nil
}

func (mgr *ContainerManager) releaseContainerResources(c *Container) error {
	mgr.releaseContainerIOs(c.ID)
	return mgr.releaseContainerNetwork(c)
}

// releaseContainerNetwork release container network when container exits or is stopped.
func (mgr *ContainerManager) releaseContainerNetwork(c *Container) error {
	c.Lock()
	defer c.Unlock()

	// NetworkMgr is nil, which means the pouch daemon is initializing.
	// And the libnetwork will also initialize, which will release all
	// staled network resources(endpoint, network and namespace). So we
	// don't need release the network resources.
	if mgr.NetworkMgr == nil {
		return nil
	}

	if c.NetworkSettings == nil {
		return nil
	}

	for name, epConfig := range c.NetworkSettings.Networks {
		endpoint := mgr.buildContainerEndpoint(c)
		endpoint.Name = name
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

// releaseContainerIOs releases container IO resources.
func (mgr *ContainerManager) releaseContainerIOs(containerID string) {
	// release resource
	io := mgr.IOs.Get(containerID)
	if io == nil {
		return
	}

	io.Close()
	mgr.IOs.Remove(containerID)
	return
}

// buildContainerEndpoint builds Endpoints according to container
// caller should lock container when calling this func.
func (mgr *ContainerManager) buildContainerEndpoint(c *Container) *networktypes.Endpoint {
	ep := BuildContainerEndpoint(c)

	if mgr.containerPlugin != nil {
		ep.Priority, ep.DisableResolver, ep.GenericParams = mgr.containerPlugin.PreCreateEndpoint(c.ID, c.Config.Env)
	}

	return ep
}

// setBaseFS keeps container basefs in meta.
func (mgr *ContainerManager) setBaseFS(ctx context.Context, c *Container, id string) {
	info, err := mgr.Client.GetSnapshot(ctx, id)
	if err != nil {
		logrus.Infof("failed to get container %s snapshot", id)
		return
	}

	// io.containerd.runtime.v1.linux as a const used by runc
	c.Lock()
	c.BaseFS = filepath.Join(mgr.Config.HomeDir, "containerd/state", "io.containerd.runtime.v1.linux", mgr.Config.Namespace, info.Name, "rootfs")
	c.Unlock()
}

// execProcessGC cleans unused exec processes config every 5 minutes.
func (mgr *ContainerManager) execProcessGC() {
	for range time.Tick(time.Duration(GCExecProcessTick) * time.Minute) {
		execProcesses := mgr.ExecProcesses.Values()
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
