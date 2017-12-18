package mgr

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/spec"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/randomid"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//ContainerMgr as an interface defines all operations against container.
type ContainerMgr interface {
	// Create a new container.
	Create(ctx context.Context, name string, config *types.ContainerConfigWrapper) (*types.ContainerCreateResp, error)

	// Start a container.
	Start(ctx context.Context, id, detachKeys string) error

	// Stop a container.
	Stop(ctx context.Context, name string, timeout time.Duration) error

	// Pause a container.
	Pause(ctx context.Context, name string) error

	// Unpause a container.
	Unpause(ctx context.Context, name string) error

	// Attach a container.
	Attach(ctx context.Context, name string, attach *AttachConfig) error

	// List returns the list of containers.
	List(ctx context.Context) ([]*types.ContainerInfo, error)

	// CreateExec creates exec process's environment.
	CreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (string, error)

	// StartExec executes a new process in container.
	StartExec(ctx context.Context, execid string, config *types.ExecStartConfig, attach *AttachConfig) error

	// Remove removes a container, it may be running or stopped and so on.
	Remove(ctx context.Context, name string, option *ContainerRemoveOption) error

	// Rename renames a container
	Rename(ctx context.Context, oldName string, newName string) error

	// Get the detailed information of container
	Get(name string) (*types.ContainerInfo, error)
}

// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
	Store         *meta.Store
	Client        *ctrd.Client
	NameToID      *collect.SafeMap
	ImageMgr      ImageMgr
	VolumeMgr     VolumeMgr
	IOs           *containerio.Cache
	ExecProcesses *collect.SafeMap
	Config        *config.Config

	// cache stores all containers in memory.
	cache *collect.SafeMap
}

// NewContainerManager creates a brand new container manager.
func NewContainerManager(ctx context.Context, store *meta.Store, cli *ctrd.Client, imgMgr ImageMgr, volMgr VolumeMgr, cfg *config.Config) (*ContainerManager, error) {
	mgr := &ContainerManager{
		Store:         store,
		NameToID:      collect.NewSafeMap(),
		Client:        cli,
		ImageMgr:      imgMgr,
		VolumeMgr:     volMgr,
		IOs:           containerio.NewCache(),
		ExecProcesses: collect.NewSafeMap(),
		cache:         collect.NewSafeMap(),
		Config:        cfg,
	}

	mgr.Client.SetStopHooks(mgr.stoppedAndRelease)
	mgr.Client.SetExitHooks(mgr.exitedAndRelease)

	return mgr, mgr.Restore(ctx)
}

// Restore containers from meta store to memory and recover those container.
func (mgr *ContainerManager) Restore(ctx context.Context) error {
	fn := func(obj meta.Object) error {
		c, ok := obj.(*types.ContainerInfo)
		if !ok {
			// object has not type of ContainerInfo
			return nil
		}

		// map container's name to id.
		mgr.NameToID.Put(c.Name, c.ID)

		// put container into cache.
		mgr.cache.Put(c.ID, &Container{
			meta: c,
		})

		if c.Status != types.StatusRunning {
			return nil
		}

		// recover the running container.
		io, err := mgr.openContainerIO(c.ID, nil)
		if err != nil {
			logrus.Errorf("failed to recover container: %s,  %v", c.ID, err)
		}

		if err := mgr.Client.RecoverContainer(ctx, c.ID, io); err == nil {
			return nil
		}

		logrus.Errorf("failed to recover container: %s,  %v", c.ID, err)
		// release io
		io.Stdin.Close()
		io.Stdout.Close()
		io.Stderr.Close()
		mgr.IOs.Remove(c.ID)

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

	if !c.IsStopped() && !c.IsCreated() && !option.Force {
		return fmt.Errorf("container: %s is not stopped, can't remove it without flag force", c.ID())
	}

	// if the container is running, force to stop it.
	if c.IsRunning() && option.Force {
		if _, err := mgr.Client.DestroyContainer(ctx, c.ID()); err != nil && !errtypes.IsNotfound(err) {
			return errors.Wrapf(err, "failed to remove container: %s", c.ID())
		}
	}

	// remove name
	mgr.NameToID.Remove(c.Name())

	// remove meta data
	if err := mgr.Store.Remove(c.meta.Key()); err != nil {
		logrus.Errorf("failed to remove container: %s meta store", c.ID())
	}

	// remove container cache
	mgr.cache.Remove(c.ID())

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

	process := &specs.Process{
		Args:     execConfig.Cmd,
		Terminal: execConfig.Tty,
		Cwd:      "/",
	}

	return mgr.Client.ExecContainer(ctx, &ctrd.Process{
		ContainerID: execConfig.ContainerID,
		ExecID:      execid,
		IO:          io,
		P:           process,
	})
}

// Create checks passed in parameters and create a Container object whose status is set at Created.
func (mgr *ContainerManager) Create(ctx context.Context, name string, config *types.ContainerConfigWrapper) (*types.ContainerCreateResp, error) {
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
	if err := mgr.parseVolumes(ctx, config); err != nil {
		return nil, errors.Wrap(err, "failed to parse volume argument")
	}

	// check the image existed or not, and convert image id to image ref
	image, err := mgr.ImageMgr.GetImage(ctx, config.Image)
	if err != nil {
		return nil, err
	}

	// set container hostconfig
	config.Image = image.Name

	if config.HostConfig.Runtime == "" {
		config.HostConfig.Runtime = mgr.Config.DefaultRuntime
	}

	// TODO add more validation of parameter
	// TODO check whether image exist
	meta := &types.ContainerInfo{
		ContainerState: &types.ContainerState{
			Status:    types.StatusCreated,
			StartedAt: strconv.FormatInt(time.Now().Unix(), 10),
		},
		ID:     id,
		Name:   name,
		Config: config,
	}

	c := &Container{
		meta: meta,
	}

	c.Lock()
	defer c.Unlock()

	// store disk
	c.Write(mgr.Store)

	// add to collection
	mgr.NameToID.Put(name, id)
	mgr.cache.Put(id, c)

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

	if c.meta.Config == nil || c.meta.ContainerState == nil {
		return errors.Wrap(errtypes.ErrNotfound, "container "+c.ID())
	}
	c.meta.DetachKeys = detachKeys

	// new a default spec.
	s, err := ctrd.NewDefaultSpec(ctx, c.ID())
	if err != nil {
		return errors.Wrapf(err, "failed to generate spec: %s", c.ID())
	}

	for _, f := range spec.SetupFuncs() {
		if err = f(ctx, c.meta, s); err != nil {
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
		Info: c.meta,
		Spec: s,
		IO:   io,
	})
	if err == nil {
		c.meta.Status = types.StatusRunning
		c.meta.StartedAt = time.Now().String()
		pid, err := mgr.Client.ContainerPID(ctx, c.ID())
		if err != nil {
			return errors.Wrapf(err, "failed to get PID of container: %s", c.ID())
		}
		c.meta.Pid = int64(pid)
	} else {
		c.meta.FinishedAt = time.Now().String()
		c.meta.Error = err.Error()
		c.meta.Pid = 0
		//TODO get and set exit code

		// release io
		io.Stderr.Close()
		io.Stdout.Close()
		io.Stdin.Close()
		mgr.IOs.Remove(c.ID())
	}

	c.Write(mgr.Store)
	return err
}

// Stop stops a running container.
func (mgr *ContainerManager) Stop(ctx context.Context, name string, timeout time.Duration) error {
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
		return fmt.Errorf("container's status is not running: %v", c.meta.Status)
	}

	if _, err := mgr.Client.DestroyContainer(ctx, c.ID()); err != nil {
		return errors.Wrapf(err, "failed to destroy container: %s", c.ID())
	}

	return nil
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

	if c.meta.Config == nil || c.meta.ContainerState == nil {
		return errors.Wrap(errtypes.ErrNotfound, "container "+c.ID())
	}

	if !c.IsRunning() {
		return fmt.Errorf("container's status is not running: %v", c.meta.Status)
	}

	if err := mgr.Client.PauseContainer(ctx, c.ID()); err != nil {
		return errors.Wrapf(err, "failed to pause container: %s", c.ID())
	}

	c.meta.Status = types.StatusPaused
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

	if c.meta.Config == nil || c.meta.ContainerState == nil {
		return errors.Wrap(errtypes.ErrNotfound, "container "+c.ID())
	}

	if !c.IsPaused() {
		return fmt.Errorf("container's status is not paused: %v", c.meta.Status)
	}

	if err := mgr.Client.UnpauseContainer(ctx, c.ID()); err != nil {
		return errors.Wrapf(err, "failed to unpause container: %s", c.ID())
	}

	c.meta.Status = types.StatusRunning
	c.Write(mgr.Store)
	return nil
}

// Attach attachs a container's io.
func (mgr *ContainerManager) Attach(ctx context.Context, name string, attach *AttachConfig) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	if _, err := mgr.openContainerIO(c.ID(), attach); err != nil {
		return err
	}

	return nil
}

// List returns the container's list.
func (mgr *ContainerManager) List(ctx context.Context) ([]*types.ContainerInfo, error) {
	cis := []*types.ContainerInfo{}

	list, err := mgr.Store.List()
	if err != nil {
		return nil, err
	}

	for _, obj := range list {
		ci, ok := obj.(*types.ContainerInfo)
		if !ok {
			return nil, fmt.Errorf("failed to get container list, invalid meta type")
		}
		cis = append(cis, ci)
	}

	return cis, nil
}

// Get the detailed information of container
func (mgr *ContainerManager) Get(name string) (*types.ContainerInfo, error) {
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

// containerInfo returns the 'ContainerInfo' object, the parameter 's' may be container's
// name, id or prefix id.
func (mgr *ContainerManager) containerInfo(s string) (*types.ContainerInfo, error) {
	var (
		obj meta.Object
		err error
	)

	// name is the container's name.
	id, ok := mgr.NameToID.Get(s).String()
	if ok {
		if obj, err = mgr.Store.Get(id); err != nil {
			return nil, errors.Wrapf(err, "failed to get container info: %s", s)
		}
	} else {
		// name is the container's prefix of the id.
		objs, err := mgr.Store.GetWithPrefix(s)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get container info with prefix: %s", s)
		}
		if len(objs) != 1 {
			return nil, fmt.Errorf("failed to get container info with prefix: %s, there are %d containers", s, len(objs))
		}
		obj = objs[0]
	}

	ci, ok := obj.(*types.ContainerInfo)
	if !ok {
		return nil, fmt.Errorf("failed to get container info, invalid meta's type")
	}

	return ci, nil
}

func (mgr *ContainerManager) openContainerIO(id string, attach *AttachConfig) (*containerio.IO, error) {
	return mgr.openIO(id, attach, false)
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
		options = append(options, containerio.WithHijack(attach.Hijack, attach.Upgrade))
		if attach.Stdin {
			options = append(options, containerio.WithStdinHijack())
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

func (mgr *ContainerManager) stoppedAndRelease(id string, m *ctrd.Message) error {
	// update container info
	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	c.meta.Pid = -1
	c.meta.ExitCode = int64(m.ExitCode())
	c.meta.FinishedAt = time.Now().String()
	c.meta.Status = types.StatusStopped

	if m.HasError() {
		c.meta.Error = m.Error().Error()
	}

	// release resource
	if io := mgr.IOs.Get(id); io != nil {
		io.Stderr.Close()
		io.Stdout.Close()
		io.Stdin.Close()
		mgr.IOs.Remove(id)
	}

	// lookup again, avoid container already removed
	if _, err := mgr.container(id); err != nil {
		if errtypes.IsNotfound(err) {
			err = nil
		}
		return err
	}

	// update meta
	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update meta: %v", err)
	}
	return nil
}

func (mgr *ContainerManager) exitedAndRelease(id string, m *ctrd.Message) error {
	if io := mgr.IOs.Get(id); io != nil {
		if err := m.StartError(); err != nil {
			fmt.Fprintf(io.Stdout, "%v\n", err)
		}

		// close io
		io.Stderr.Close()
		io.Stdout.Close()
		io.Stdin.Close()
		mgr.IOs.Remove(id)
	}
	mgr.ExecProcesses.Remove(id)
	return nil
}

func (mgr *ContainerManager) parseVolumes(ctx context.Context, c *types.ContainerConfigWrapper) error {
	logrus.Debugf("bind volumes: %v", c.HostConfig.Binds)
	// TODO: parse c.HostConfig.VolumesFrom

	for i, b := range c.HostConfig.Binds {
		// TODO: when caused error, how to rollback.
		arr, err := checkBind(b)
		if err != nil {
			return err
		}
		source := ""
		switch len(arr) {
		case 1:
			source = ""
		case 2, 3:
			source = arr[0]
		default:
			return errors.Errorf("unknown bind: %s", b)
		}

		if source == "" {
			source = randomid.Generate()
		}
		if !path.IsAbs(source) {
			_, err := mgr.VolumeMgr.Info(ctx, source)
			if err != nil {
				opts := map[string]string{
					"backend": "local",
					"size":    "100G",
				}
				if err := mgr.VolumeMgr.Create(ctx, source, c.HostConfig.VolumeDriver, opts, nil); err != nil {
					logrus.Errorf("failed to create volume: %s, err: %v", source, err)
					return errors.Wrap(err, "failed to create volume")
				}
			}

			if _, err := mgr.VolumeMgr.Attach(ctx, source, nil); err != nil {
				logrus.Errorf("failed to attach volume: %s, err: %v", source, err)
				return errors.Wrap(err, "failed to attach volume")
			}

			mountPath, err := mgr.VolumeMgr.Path(ctx, source)
			if err != nil {
				logrus.Errorf("failed to get the mount path of volume: %s, err: %v", source, err)
				return errors.Wrap(err, "failed to get volume mount path")
			}

			source = mountPath
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
