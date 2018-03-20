package mgr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/lxcfs"
	networktypes "github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/randomid"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/containerd/containerd/namespaces"

	"github.com/go-openapi/strfmt"
	"github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ContainerMgr as an interface defines all operations against container.
type ContainerMgr interface {
	// Create a new container.
	Create(ctx context.Context, name string, config *types.ContainerCreateConfig) (*types.ContainerCreateResp, error)

	// Start a container.
	Start(ctx context.Context, id, detachKeys string) error

	// Stop a container.
	Stop(ctx context.Context, name string, timeout int64) error

	// Pause a container.
	Pause(ctx context.Context, name string) error

	// Unpause a container.
	Unpause(ctx context.Context, name string) error

	// Attach a container.
	Attach(ctx context.Context, name string, attach *AttachConfig) error

	// List returns the list of containers.
	List(ctx context.Context, filter ContainerFilter, option *ContainerListOption) ([]*ContainerMeta, error)

	// CreateExec creates exec process's environment.
	CreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (string, error)

	// StartExec executes a new process in container.
	StartExec(ctx context.Context, execid string, config *types.ExecStartConfig, attach *AttachConfig) error

	// InspectExec returns low-level information about exec command.
	InspectExec(ctx context.Context, execid string) (*ContainerExecInspect, error)

	// Remove removes a container, it may be running or stopped and so on.
	Remove(ctx context.Context, name string, option *ContainerRemoveOption) error

	// Rename renames a container.
	Rename(ctx context.Context, oldName string, newName string) error

	// Get the detailed information of container.
	Get(ctx context.Context, name string) (*ContainerMeta, error)

	// Update updates the configurations of a container.
	Update(ctx context.Context, name string, config *types.UpdateConfig) error

	// Upgrade upgrades a container with new image and args.
	Upgrade(ctx context.Context, name string, config *types.ContainerUpgradeConfig) error

	// Top lists the processes running inside of the given container
	Top(ctx context.Context, name string, psArgs string) (*types.ContainerProcessList, error)

	// Resize resizes the size of container tty.
	Resize(ctx context.Context, name string, opts types.ResizeOptions) error
}

// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
	// Store stores containers in Backend store.
	// Element operated in store must has a type of *ContainerMeta.
	// By default, Store will use local filesystem with json format to store containers.
	Store *meta.Store

	// Client is used to interact with containerd.
	Client *ctrd.Client

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
}

// NewContainerManager creates a brand new container manager.
func NewContainerManager(ctx context.Context, store *meta.Store, cli *ctrd.Client, imgMgr ImageMgr, volMgr VolumeMgr, netMgr NetworkMgr, cfg *config.Config) (*ContainerManager, error) {
	mgr := &ContainerManager{
		Store:         store,
		NameToID:      collect.NewSafeMap(),
		Client:        cli,
		ImageMgr:      imgMgr,
		VolumeMgr:     volMgr,
		NetworkMgr:    netMgr,
		IOs:           containerio.NewCache(),
		ExecProcesses: collect.NewSafeMap(),
		cache:         collect.NewSafeMap(),
		Config:        cfg,
		monitor:       NewContainerMonitor(),
	}

	mgr.Client.SetExitHooks(mgr.exitedAndRelease)
	mgr.Client.SetExecExitHooks(mgr.execExitedAndRelease)

	return mgr, mgr.Restore(ctx)
}

// Restore containers from meta store to memory and recover those container.
func (mgr *ContainerManager) Restore(ctx context.Context) error {
	fn := func(obj meta.Object) error {
		containerMeta, ok := obj.(*ContainerMeta)
		if !ok {
			// object has not type of ContainerMeta
			return nil
		}

		// map container's name to id.
		mgr.NameToID.Put(containerMeta.Name, containerMeta.ID)

		// put container into cache.
		mgr.cache.Put(containerMeta.ID, &Container{meta: containerMeta})

		if containerMeta.State.Status != types.StatusRunning {
			return nil
		}

		// recover the running container.
		io, err := mgr.openContainerIO(containerMeta.ID, nil)
		if err != nil {
			logrus.Errorf("failed to recover container: %s,  %v", containerMeta.ID, err)
		}

		if err := mgr.Client.RecoverContainer(ctx, containerMeta.ID, io); err == nil {
			return nil
		}

		logrus.Errorf("failed to recover container: %s,  %v", containerMeta.ID, err)
		// release io
		io.Close()
		mgr.IOs.Remove(containerMeta.ID)

		return nil
	}
	return mgr.Store.ForEach(fn)
}

// Remove removes a container, it may be running or stopped and so on.
func (mgr *ContainerManager) Remove(ctx context.Context, name string, option *ContainerRemoveOption) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}
	c.Lock()
	defer c.Unlock()

	if !c.IsStopped() && !c.IsExited() && !c.IsCreated() && !option.Force {
		return fmt.Errorf("container: %s is not stopped, can't remove it without flag force", c.ID())
	}

	// if the container is running, force to stop it.
	if c.IsRunning() && option.Force {
		msg, err := mgr.Client.DestroyContainer(ctx, c.ID(), c.StopTimeout())
		if err != nil && !errtypes.IsNotfound(err) {
			return errors.Wrapf(err, "failed to destory container: %s", c.ID())
		}
		if err := mgr.markStoppedAndRelease(c, msg); err != nil {
			return errors.Wrapf(err, "failed to mark container: %s stop status", c.ID())
		}
	}

	if err := mgr.detachVolumes(ctx, c.meta); err != nil {
		logrus.Errorf("failed to detach volume: %v", err)
	}

	// remove name
	mgr.NameToID.Remove(c.Name())

	// remove meta data
	if err := mgr.Store.Remove(c.meta.Key()); err != nil {
		logrus.Errorf("failed to remove container: %s meta store, %v", c.ID(), err)
	}

	// remove container cache
	mgr.cache.Remove(c.ID())

	// remove snapshot
	if err := mgr.Client.RemoveSnapshot(ctx, c.ID()); err != nil {
		logrus.Errorf("failed to remove container: %s snapshot, %v", c.ID(), err)
	}

	return nil
}

// CreateExec creates exec process's meta data.
func (mgr *ContainerManager) CreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (string, error) {
	c, err := mgr.container(name)
	if err != nil {
		return "", err
	}

	if !c.IsRunning() {
		return "", fmt.Errorf("container %s is not running", c.ID())
	}

	execid := randomid.Generate()
	execConfig := &containerExecConfig{
		ExecCreateConfig: *config,
		ContainerID:      c.ID(),
	}
	execConfig.exitCh = make(chan *ctrd.Message, 1)

	mgr.ExecProcesses.Put(execid, execConfig)

	return execid, nil
}

// StartExec executes a new process in container.
func (mgr *ContainerManager) StartExec(ctx context.Context, execid string, config *types.ExecStartConfig, attach *AttachConfig) error {
	v, ok := mgr.ExecProcesses.Get(execid).Result()
	if !ok {
		return errors.Wrap(errtypes.ErrNotfound, "to be exec process: "+execid)
	}
	execConfig, ok := v.(*containerExecConfig)
	if !ok {
		return fmt.Errorf("invalid exec config type")
	}

	if attach != nil {
		attach.Stdin = execConfig.AttachStdin
	}

	io, err := mgr.openExecIO(execid, attach)
	if err != nil {
		return err
	}

	c, err := mgr.container(execConfig.ContainerID)
	if err != nil {
		return err
	}

	process := &specs.Process{
		Args:     execConfig.Cmd,
		Terminal: execConfig.Tty,
		Cwd:      "/",
		Env:      c.Config().Env,
	}

	if execConfig.User != "" {
		c.meta.Config.User = execConfig.User
	}

	if err = setupProcessUser(ctx, c.meta, &SpecWrapper{s: &specs.Spec{Process: process}}); err != nil {
		return err
	}

	return mgr.Client.ExecContainer(ctx, &ctrd.Process{
		ContainerID: execConfig.ContainerID,
		ExecID:      execid,
		IO:          io,
		P:           process,
	})
}

// InspectExec returns low-level information about exec command.
func (mgr *ContainerManager) InspectExec(ctx context.Context, execid string) (*ContainerExecInspect, error) {
	v, ok := mgr.ExecProcesses.Get(execid).Result()
	if !ok {
		return nil, errors.Wrap(errtypes.ErrNotfound, "to be exec process: "+execid)
	}
	execConfig, ok := v.(*containerExecConfig)
	if !ok {
		return nil, fmt.Errorf("invalid exec config type")
	}

	return &ContainerExecInspect{
		ExitCh: execConfig.exitCh,
	}, nil
}

// Create checks passed in parameters and create a Container object whose status is set at Created.
func (mgr *ContainerManager) Create(ctx context.Context, name string, config *types.ContainerCreateConfig) (*types.ContainerCreateResp, error) {
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

	// parse volume config
	if err := mgr.parseVolumes(ctx, id, config); err != nil {
		return nil, errors.Wrap(err, "failed to parse volume argument")
	}

	// check the image existed or not, and convert image id to image ref
	image, err := mgr.ImageMgr.GetImage(ctx, config.Image)
	if err != nil {
		return nil, err
	}
	config.Image = image.Name

	// set container runtime
	if config.HostConfig.Runtime == "" {
		config.HostConfig.Runtime = mgr.Config.DefaultRuntime
	}

	// create a snapshot with image.
	if err := mgr.Client.CreateSnapshot(ctx, id, config.Image); err != nil {
		return nil, err
	}

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

	meta := &ContainerMeta{
		State: &types.ContainerState{
			Status: types.StatusCreated,
		},
		ID:         id,
		Image:      image.Name,
		Name:       name,
		Config:     &config.ContainerConfig,
		Created:    time.Now().UTC().Format(utils.TimeLayout),
		HostConfig: config.HostConfig,
	}

	// set container basefs
	mgr.setBaseFS(ctx, meta, id)

	// set network settings
	networkMode := config.HostConfig.NetworkMode
	if networkMode == "" {
		config.HostConfig.NetworkMode = "bridge"
		meta.Config.NetworkDisabled = true
	}
	meta.NetworkSettings = new(types.NetworkSettings)
	if len(config.NetworkingConfig.EndpointsConfig) > 0 {
		meta.NetworkSettings.Networks = config.NetworkingConfig.EndpointsConfig
	}
	if meta.NetworkSettings.Networks == nil && networkMode != "" && !IsContainer(networkMode) {
		meta.NetworkSettings.Networks = make(map[string]*types.EndpointSettings)
		meta.NetworkSettings.Networks[config.HostConfig.NetworkMode] = new(types.EndpointSettings)
	}

	if err := parseSecurityOpts(meta, config.HostConfig.SecurityOpt); err != nil {
		return nil, err
	}

	// merge image's config into container's meta
	if err := meta.merge(func() (v1.ImageConfig, error) {
		ociimage, err := mgr.Client.GetOciImage(ctx, config.Image)
		if err != nil {
			return ociimage.Config, err
		}
		return ociimage.Config, nil
	}); err != nil {
		return nil, err
	}

	container := &Container{
		meta: meta,
	}

	container.Lock()
	defer container.Unlock()

	// store disk
	container.Write(mgr.Store)

	// add to collection
	mgr.NameToID.Put(name, id)
	mgr.cache.Put(id, container)

	return &types.ContainerCreateResp{
		ID:   id,
		Name: name,
	}, nil
}

// Start a pre created Container.
func (mgr *ContainerManager) Start(ctx context.Context, id, detachKeys string) (err error) {
	if id == "" {
		return errors.Wrap(errtypes.ErrInvalidParam, "either container name or id is required")
	}

	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if c.meta.Config == nil || c.meta.State == nil {
		return errors.Wrap(errtypes.ErrNotfound, "container "+c.ID())
	}
	c.DetachKeys = detachKeys

	// initialise container network mode
	networkMode := c.meta.HostConfig.NetworkMode
	if IsContainer(networkMode) {
		origContainer, err := mgr.Get(ctx, strings.SplitN(networkMode, ":", 2)[1])
		if err != nil {
			return err
		}

		c.meta.HostnamePath = origContainer.HostnamePath
		c.meta.HostsPath = origContainer.HostsPath
		c.meta.ResolvConfPath = origContainer.ResolvConfPath
		c.meta.Config.Hostname = origContainer.Config.Hostname
		c.meta.Config.Domainname = origContainer.Config.Domainname
	}

	// initialise host network mode
	if IsHost(networkMode) {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		c.meta.Config.Hostname = strfmt.Hostname(hostname)
	}

	// initialise network endpoint
	if c.meta.NetworkSettings != nil {
		for name, endpointSetting := range c.meta.NetworkSettings.Networks {
			endpoint := mgr.buildContainerEndpoint(c.meta)
			endpoint.Name = name
			endpoint.EndpointConfig = endpointSetting
			if _, err := mgr.NetworkMgr.EndpointCreate(ctx, endpoint); err != nil {
				logrus.Errorf("failed to create endpoint: %v", err)
				return err
			}
		}
	}

	// new a default spec.
	s, err := ctrd.NewDefaultSpec(ctx, c.ID())
	if err != nil {
		return errors.Wrapf(err, "failed to generate spec: %s", c.ID())
	}

	var cgroupsParent string
	if c.meta.HostConfig.CgroupParent != "" {
		cgroupsParent = c.meta.HostConfig.CgroupParent
	} else if mgr.Config.CgroupParent != "" {
		cgroupsParent = mgr.Config.CgroupParent
	}

	// cgroupsPath must be absolute path
	if cgroupsParent != "" {
		if !filepath.IsAbs(cgroupsParent) {
			cgroupsParent = filepath.Clean("/" + cgroupsParent)
		}

		s.Linux.CgroupsPath = filepath.Join(cgroupsParent, c.ID())
	}

	sw := &SpecWrapper{
		s:      s,
		ctrMgr: mgr,
		volMgr: mgr.VolumeMgr,
		netMgr: mgr.NetworkMgr,
	}

	for _, setup := range SetupFuncs() {
		if err = setup(ctx, c.meta, sw); err != nil {
			return err
		}
	}

	// open container's stdio.
	io, err := mgr.openContainerIO(c.ID(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to open io")
	}
	if io.Stdin.OpenStdin() {
		s.Process.Terminal = true
	}

	err = mgr.Client.CreateContainer(ctx, &ctrd.Container{
		ID:      c.ID(),
		Image:   c.Image(),
		Runtime: c.meta.HostConfig.Runtime,
		Spec:    s,
		IO:      io,
	})
	if err == nil {
		c.meta.State.Status = types.StatusRunning
		c.meta.State.StartedAt = time.Now().UTC().Format(utils.TimeLayout)
		pid, err := mgr.Client.ContainerPID(ctx, c.ID())
		if err != nil {
			return errors.Wrapf(err, "failed to get PID of container: %s", c.ID())
		}
		c.meta.State.Pid = int64(pid)
	} else {
		c.meta.State.FinishedAt = time.Now().UTC().Format(utils.TimeLayout)
		c.meta.State.Error = err.Error()
		c.meta.State.Pid = 0
		//TODO: make exit code more correct.
		c.meta.State.ExitCode = 127

		// release io
		io.Close()
		mgr.IOs.Remove(c.ID())
	}

	c.Write(mgr.Store)
	return err
}

// Stop stops a running container.
func (mgr *ContainerManager) Stop(ctx context.Context, name string, timeout int64) error {
	var (
		err error
		c   *Container
	)

	if c, err = mgr.container(name); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if !c.IsRunning() {
		return fmt.Errorf("container's status is not running: %s", c.meta.State.Status)
	}

	if timeout == 0 {
		timeout = c.StopTimeout()
	}

	msg, err := mgr.Client.DestroyContainer(ctx, c.ID(), timeout)
	if err != nil {
		return errors.Wrapf(err, "failed to destroy container: %s", c.ID())
	}

	return mgr.markStoppedAndRelease(c, msg)
}

// Pause pauses a running container.
func (mgr *ContainerManager) Pause(ctx context.Context, name string) error {
	var (
		err error
		c   *Container
	)

	if c, err = mgr.container(name); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if c.meta.Config == nil || c.meta.State == nil {
		return errors.Wrap(errtypes.ErrNotfound, "container "+c.ID())
	}

	if !c.IsRunning() {
		return fmt.Errorf("container's status is not running: %s", c.meta.State.Status)
	}

	if err := mgr.Client.PauseContainer(ctx, c.ID()); err != nil {
		return errors.Wrapf(err, "failed to pause container: %s", c.ID())
	}

	c.meta.State.Status = types.StatusPaused
	c.Write(mgr.Store)
	return nil
}

// Unpause unpauses a paused container.
func (mgr *ContainerManager) Unpause(ctx context.Context, name string) error {
	var (
		err error
		c   *Container
	)

	if c, err = mgr.container(name); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if c.meta.Config == nil || c.meta.State == nil {
		return errors.Wrap(errtypes.ErrNotfound, "container "+c.ID())
	}

	if !c.IsPaused() {
		return fmt.Errorf("container's status is not paused: %v", c.meta.State.Status)
	}

	if err := mgr.Client.UnpauseContainer(ctx, c.ID()); err != nil {
		return errors.Wrapf(err, "failed to unpause container: %s", c.ID())
	}

	c.meta.State.Status = types.StatusRunning
	c.Write(mgr.Store)
	return nil
}

// Attach attachs a container's io.
func (mgr *ContainerManager) Attach(ctx context.Context, name string, attach *AttachConfig) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	_, err = mgr.openAttachIO(c.ID(), attach)
	if err != nil {
		return err
	}

	return nil
}

// List returns the container's list.
func (mgr *ContainerManager) List(ctx context.Context, filter ContainerFilter, option *ContainerListOption) ([]*ContainerMeta, error) {
	metas := []*ContainerMeta{}

	list, err := mgr.Store.List()
	if err != nil {
		return nil, err
	}

	for _, obj := range list {
		m, ok := obj.(*ContainerMeta)
		if !ok {
			return nil, fmt.Errorf("failed to get container list, invalid meta type")
		}
		if filter != nil && filter(m) {
			if option.All {
				metas = append(metas, m)
			} else if m.State.Status == types.StatusRunning || m.State.Status == types.StatusPaused {
				metas = append(metas, m)
			}
		}
	}

	return metas, nil
}

// Get the detailed information of container.
func (mgr *ContainerManager) Get(ctx context.Context, name string) (*ContainerMeta, error) {
	c, err := mgr.container(name)
	if err != nil {
		return nil, err
	}

	c.Lock()
	defer c.Unlock()

	return c.meta, nil
}

// Rename renames a container
func (mgr *ContainerManager) Rename(ctx context.Context, oldName, newName string) error {
	var (
		c   *Container
		err error
	)

	if mgr.NameToID.Get(newName).Exist() {
		return errors.Wrap(errtypes.ErrAlreadyExisted, "container name: "+newName)
	}

	if c, err = mgr.container(oldName); err != nil {
		return errors.Wrap(err, "failed to rename container")
	}
	c.Lock()
	defer c.Unlock()

	mgr.NameToID.Remove(oldName)
	mgr.NameToID.Put(newName, c.ID())

	c.meta.Name = newName
	c.Write(mgr.Store)

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

	// update ContainerConfig of a container.
	if !c.IsStopped() && config.Image != "" || len(config.Env) > 0 {
		return fmt.Errorf("Only can update the container's image or Env when it is stopped")
	}

	if config.Image != "" {
		image, err := mgr.ImageMgr.GetImage(ctx, config.Image)
		if err != nil {
			return err
		}
		// TODO Image param is duplicate in ContainerMeta
		c.meta.Config.Image = image.Name
		c.meta.Image = image.Name
	}

	if len(config.Env) != 0 {
		for k, v := range config.Env {
			c.meta.Config.Env[k] = v
		}
	}

	if len(config.Labels) != 0 {
		for k, v := range config.Labels {
			c.meta.Config.Labels[k] = v
		}
	}

	// update resources of container.
	resources := config.Resources
	cResources := &c.meta.HostConfig.Resources
	if resources.BlkioWeight != 0 {
		cResources.BlkioWeight = resources.BlkioWeight
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
		cResources.Memory = resources.Memory
	}
	if resources.MemorySwap != 0 {
		cResources.MemorySwap = resources.MemorySwap
	}

	// update HostConfig of a container.
	// TODO update restartpolicy when container is running.
	if config.RestartPolicy.Name != "" {
		c.meta.HostConfig.RestartPolicy = config.RestartPolicy
	}

	// If container is not running, update container metadata struct is enough,
	// resources will be updated when the container is started again,
	// If container is running, we need to update configs to the real world.
	var updateErr error
	if c.IsRunning() {
		updateErr = mgr.Client.UpdateResources(ctx, c.ID(), c.meta.HostConfig.Resources)
	}

	// store disk.
	if updateErr == nil {
		c.Write(mgr.Store)
	}

	return updateErr
}

// Upgrade upgrades a container with new image and args.
func (mgr *ContainerManager) Upgrade(ctx context.Context, name string, config *types.ContainerUpgradeConfig) error {
	// TODO
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
		return nil, fmt.Errorf("container is not running, can not execute top command")
	}

	pids, err := mgr.Client.GetPidsForContainer(ctx, c.ID())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get pids of container")
	}

	output, err := exec.Command("ps", strings.Split(psArgs, " ")...).Output()
	if err != nil {
		return nil, errors.Wrapf(err, "error running ps")
	}

	procList, err := parsePSOutput(output, pids)
	if err != nil {
		return nil, errors.Wrapf(err, "parsePSOutput failed")
	}

	return procList, nil
}

// Resize resizes the size of a container tty.
func (mgr *ContainerManager) Resize(ctx context.Context, name string, opts types.ResizeOptions) error {
	// TODO
	return nil
}

func (mgr *ContainerManager) openContainerIO(id string, attach *AttachConfig) (*containerio.IO, error) {
	return mgr.openIO(id, attach, false)
}

func (mgr *ContainerManager) openAttachIO(id string, attach *AttachConfig) (*containerio.IO, error) {
	io := mgr.IOs.Get(id)
	if io == nil {
		return mgr.openIO(id, attach, false)
	}

	options := []func(*containerio.Option){
		containerio.WithID(id),
	}
	if attach != nil {
		if attach.Hijack != nil {
			// Attaching using http.
			options = append(options, containerio.WithHijack(attach.Hijack, attach.Upgrade))
			if attach.Stdin {
				options = append(options, containerio.WithStdinHijack())
			}
		} else if attach.MemBuffer != nil {
			// Attaching using memory buffer.
			options = append(options, containerio.WithMemBuffer(attach.MemBuffer))
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
	} else {
		options = append(options, containerio.WithDiscard())
	}

	io.AddBackend(containerio.NewOption(options...))

	mgr.IOs.Put(id, io)

	return io, nil
}

func (mgr *ContainerManager) openExecIO(id string, attach *AttachConfig) (*containerio.IO, error) {
	return mgr.openIO(id, attach, true)
}

func (mgr *ContainerManager) openIO(id string, attach *AttachConfig, exec bool) (*containerio.IO, error) {
	if io := mgr.IOs.Get(id); io != nil {
		return io, nil
	}

	options := []func(*containerio.Option){
		containerio.WithID(id),
	}

	if !exec {
		root := mgr.Store.Path(id)
		options = append(options, containerio.WithRootDir(root))
	}

	if attach != nil {
		if attach.Hijack != nil {
			// Attaching using http.
			options = append(options, containerio.WithHijack(attach.Hijack, attach.Upgrade))
			if attach.Stdin {
				options = append(options, containerio.WithStdinHijack())
			}
		} else if attach.MemBuffer != nil {
			// Attaching using memory buffer.
			options = append(options, containerio.WithMemBuffer(attach.MemBuffer))
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
	} else if !exec {
		options = append(options, containerio.WithRawFile())

	} else {
		options = append(options, containerio.WithDiscard())
	}

	io := containerio.NewIO(containerio.NewOption(options...))

	mgr.IOs.Put(id, io)

	return io, nil
}

func (mgr *ContainerManager) markStoppedAndRelease(c *Container, m *ctrd.Message) error {
	c.meta.State.Pid = -1
	c.meta.State.FinishedAt = time.Now().UTC().Format(utils.TimeLayout)
	c.meta.State.Status = types.StatusStopped

	if m != nil {
		c.meta.State.ExitCode = int64(m.ExitCode())
		if err := m.RawError(); err != nil {
			c.meta.State.Error = err.Error()
		}
	}

	// release resource
	if io := mgr.IOs.Get(c.ID()); io != nil {
		io.Close()
		mgr.IOs.Remove(c.ID())
	}

	// release network
	if c.meta.NetworkSettings != nil {
		for name, epConfig := range c.meta.NetworkSettings.Networks {
			endpoint := mgr.buildContainerEndpoint(c.meta)
			endpoint.Name = name
			endpoint.EndpointConfig = epConfig
			if err := mgr.NetworkMgr.EndpointRemove(context.Background(), endpoint); err != nil {
				logrus.Errorf("failed to remove endpoint: %v", err)
				return err
			}
		}
	}

	// update meta
	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update meta: %v", err)
	}
	return nil
}

// exitedAndRelease be register into ctrd as a callback function, when the running container suddenly
// exited, "ctrd" will call it to set the container's state and release resouce and so on.
func (mgr *ContainerManager) exitedAndRelease(id string, m *ctrd.Message) error {
	// update container info
	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	if err := mgr.markStoppedAndRelease(c, m); err != nil {
		return err
	}

	c.meta.State.Status = types.StatusExited
	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update meta: %v", err)
	}

	// send exit event to monitor
	mgr.monitor.PostEvent(ContainerExitEvent(c).WithHandle(func(container *Container) error {
		// check status and restart policy
		container.Lock()

		if !container.IsExited() {
			container.Unlock()
			return nil
		}
		policy := (*ContainerRestartPolicy)(container.HostConfig().RestartPolicy)
		if policy == nil || policy.IsNone() {
			container.Unlock()
			return nil
		}

		container.Unlock()

		return mgr.Start(context.TODO(), container.ID(), container.DetachKeys)
	}))

	return nil
}

// execExitedAndRelease be register into ctrd as a callback function, when the exec process in a container
// exited, "ctrd" will call it to release resource and so on.
func (mgr *ContainerManager) execExitedAndRelease(id string, m *ctrd.Message) error {
	if io := mgr.IOs.Get(id); io != nil {
		if err := m.RawError(); err != nil {
			fmt.Fprintf(io.Stdout, "%v\n", err)
		}

		// close io
		io.Close()
		mgr.IOs.Remove(id)
	}

	v, ok := mgr.ExecProcesses.Get(id).Result()
	if !ok {
		return errors.Wrap(errtypes.ErrNotfound, "to be exec process: "+id)
	}
	execConfig, ok := v.(*containerExecConfig)
	if !ok {
		return fmt.Errorf("invalid exec config type")
	}
	execConfig.exitCh <- m

	// TODO: GC invalid mgr.ExecProcess.
	return nil
}

func (mgr *ContainerManager) parseVolumes(ctx context.Context, id string, c *types.ContainerCreateConfig) error {
	logrus.Debugf("bind volumes: %v", c.HostConfig.Binds)

	if c.Volumes == nil {
		c.Volumes = make(map[string]interface{})
	}

	// TODO: parse c.HostConfig.VolumesFrom

	for i, b := range c.HostConfig.Binds {
		// TODO: when caused error, how to rollback.
		arr, err := checkBind(b)
		if err != nil {
			return err
		}
		source := ""
		destination := ""
		switch len(arr) {
		case 1:
			source = ""
			destination = arr[0]
		case 2, 3:
			source = arr[0]
			destination = arr[1]
		default:
			return errors.Errorf("unknown bind: %s", b)
		}

		if source == "" {
			source = randomid.Generate()
		}
		if !path.IsAbs(source) {
			ref := ""
			v, err := mgr.VolumeMgr.Get(ctx, source)
			if err != nil || v == nil {
				opts := map[string]string{
					"backend": "local",
				}
				if err := mgr.VolumeMgr.Create(ctx, source, c.HostConfig.VolumeDriver, opts, nil); err != nil {
					logrus.Errorf("failed to create volume: %s, err: %v", source, err)
					return errors.Wrap(err, "failed to create volume")
				}
			} else {
				ref = v.Option("ref")
			}

			option := map[string]string{}
			if ref == "" {
				option["ref"] = id
			} else {
				option["ref"] = ref + "," + id
			}
			if _, err := mgr.VolumeMgr.Attach(ctx, source, option); err != nil {
				logrus.Errorf("failed to attach volume: %s, err: %v", source, err)
				return errors.Wrap(err, "failed to attach volume")
			}

			mountPath, err := mgr.VolumeMgr.Path(ctx, source)
			if err != nil {
				logrus.Errorf("failed to get the mount path of volume: %s, err: %v", source, err)
				return errors.Wrap(err, "failed to get volume mount path")
			}

			c.Volumes[source] = destination
			source = mountPath
		} else if _, err := os.Stat(source); err != nil {
			if !os.IsNotExist(err) {
				return errors.Errorf("failed to stat %q: %v", source, err)
			}
			// Create the host path if it doesn't exist.
			if err := os.MkdirAll(source, 0755); err != nil {
				return errors.Errorf("failed to mkdir %q: %v", source, err)
			}
		}

		switch len(arr) {
		case 1:
			b = fmt.Sprintf("%s:%s", source, arr[0])
		case 2, 3:
			arr[0] = source
			b = strings.Join(arr, ":")
		default:
		}

		c.HostConfig.Binds[i] = b
	}
	return nil
}

func (mgr *ContainerManager) detachVolumes(ctx context.Context, c *ContainerMeta) error {
	for name := range c.Config.Volumes {
		v, err := mgr.VolumeMgr.Get(ctx, name)
		if err != nil {
			logrus.Errorf("failed to get volume: %s", name)
			return err
		}

		option := map[string]string{}
		ref := v.Option("ref")
		if ref == "" {
			continue
		}
		if !strings.Contains(ref, c.ID) {
			continue
		}

		ids := strings.Split(ref, ",")
		for i, id := range ids {
			if id == c.ID {
				ids = append(ids[:i], ids[i+1:]...)
				break
			}
		}
		if len(ids) > 0 {
			option["ref"] = strings.Join(ids, ",")
		} else {
			option["ref"] = ""
		}

		mgr.VolumeMgr.Detach(ctx, name, option)
	}

	return nil
}

func (mgr *ContainerManager) buildContainerEndpoint(c *ContainerMeta) *networktypes.Endpoint {
	return &networktypes.Endpoint{
		Owner:           c.ID,
		Hostname:        c.Config.Hostname,
		Domainname:      c.Config.Domainname,
		HostsPath:       c.HostsPath,
		ExtraHosts:      c.HostConfig.ExtraHosts,
		HostnamePath:    c.HostnamePath,
		ResolvConfPath:  c.ResolvConfPath,
		NetworkDisabled: c.Config.NetworkDisabled,
		NetworkMode:     c.HostConfig.NetworkMode,
		DNS:             c.HostConfig.DNS,
		DNSOptions:      c.HostConfig.DNSOptions,
		DNSSearch:       c.HostConfig.DNSSearch,
		MacAddress:      c.Config.MacAddress,
		PublishAllPorts: c.HostConfig.PublishAllPorts,
		ExposedPorts:    c.Config.ExposedPorts,
		PortBindings:    c.HostConfig.PortBindings,
		NetworkConfig:   c.NetworkSettings,
	}
}

// setBaseFS keeps container basefs in meta
func (mgr *ContainerManager) setBaseFS(ctx context.Context, meta *ContainerMeta, container string) {
	info, err := mgr.Client.GetSnapshot(ctx, container)
	if err != nil {
		logrus.Infof("failed to get container %s snapshot", container)
		return
	}

	// io.containerd.runtime.v1.linux as a const used by runc
	meta.BaseFS = filepath.Join(mgr.Config.HomeDir, "containerd/state", "io.containerd.runtime.v1.linux", namespaces.Default, info.Name, "rootfs")
}

func checkBind(b string) ([]string, error) {
	if strings.Count(b, ":") > 2 {
		return nil, fmt.Errorf("unknown volume bind: %s", b)
	}

	arr := strings.SplitN(b, ":", 3)
	switch len(arr) {
	case 1:
		if arr[0] == "" {
			return nil, fmt.Errorf("unknown volume bind: %s", b)
		}
		if arr[0][:1] != "/" {
			return nil, fmt.Errorf("invalid bind path: %s", arr[0])
		}
	case 2, 3:
		if arr[1] == "" {
			return nil, fmt.Errorf("unknown volume bind: %s", b)
		}
		if arr[1][:1] != "/" {
			return nil, fmt.Errorf("invalid bind path: %s", arr[1])
		}
	default:
		return nil, fmt.Errorf("unknown volume bind: %s", b)
	}

	return arr, nil
}
