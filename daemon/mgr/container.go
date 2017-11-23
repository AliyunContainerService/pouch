package mgr

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/spec"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/alibaba/pouch/pkg/kmutex"
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
	Start(ctx context.Context, config types.ContainerStartConfig) error

	// Stop a container.
	Stop(ctx context.Context, name string, timeout time.Duration) error

	// Attach a container.
	Attach(ctx context.Context, name string, attach *types.AttachConfig) error

	// List returns the list of containers.
	List(ctx context.Context) ([]*types.ContainerInfo, error)

	// CreateExec creates exec process's environment.
	CreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (string, error)

	// StartExec executes a new process in container.
	StartExec(ctx context.Context, execid string, config *types.ExecStartConfig, attach *types.AttachConfig) error
}

// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
	Store         *meta.Store
	Client        *ctrd.Client
	NameToID      *collect.SafeMap
	ImageMgr      ImageMgr
	VolumeMgr     VolumeMgr
	IOs           *containerio.Cache
	km            *kmutex.KMutex
	ExecProcesses *collect.SafeMap
}

// NewContainerManager creates a brand new container manager.
func NewContainerManager(ctx context.Context, store *meta.Store, cli *ctrd.Client, imgMgr ImageMgr, volMgr VolumeMgr) (*ContainerManager, error) {
	mgr := &ContainerManager{
		Store:         store,
		NameToID:      collect.NewSafeMap(),
		Client:        cli,
		ImageMgr:      imgMgr,
		VolumeMgr:     volMgr,
		IOs:           containerio.NewCache(),
		km:            kmutex.New(),
		ExecProcesses: collect.NewSafeMap(),
	}

	mgr.Client.SetStopHooks(mgr.stoppedAndRelease)
	mgr.Client.SetExitHooks(mgr.exitedAndRelease)

	return mgr, mgr.Restore(ctx)
}

// Restore containers from meta store to memory and recover those container.
func (cm *ContainerManager) Restore(ctx context.Context) error {
	fn := func(obj meta.Object) error {
		if c, ok := obj.(*types.ContainerInfo); ok {
			// map container's name to id.
			cm.NameToID.Put(c.Name, c.ID)

			// recover the running container.
			if c.Status == types.RUNNING {
				io, err := cm.openContainerIO(c.ID, nil)
				if err != nil {
					logrus.Errorf("failed to recover container: %s,  %v", c.ID, err)
				}
				if err := cm.Client.RecoverContainer(ctx, c.ID, io); err != nil {
					logrus.Errorf("failed to recover container: %s,  %v", c.ID, err)

					// release io
					io.Stdin.Close()
					io.Stdout.Close()
					io.Stderr.Close()
					cm.IOs.Remove(c.ID)
				}
			}
		}
		return nil
	}
	return cm.Store.ForEach(fn)
}

// CreateExec creates exec process's meta data.
func (cm *ContainerManager) CreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (string, error) {
	execid := randomid.Generate()

	container, err := cm.containerInfo(name)
	if err != nil {
		return "", err
	}

	execConfig := &containerExecConfig{
		ExecCreateConfig: *config,
		ContainerID:      container.ID,
	}

	cm.ExecProcesses.Put(execid, execConfig)

	return execid, nil
}

// StartExec executes a new process in container.
func (cm *ContainerManager) StartExec(ctx context.Context, execid string, config *types.ExecStartConfig, attach *types.AttachConfig) error {
	v, ok := cm.ExecProcesses.Get(execid).Result()
	if !ok {
		return fmt.Errorf("exec process: %s not found", execid)
	}
	execConfig, ok := v.(*containerExecConfig)
	if !ok {
		return fmt.Errorf("invalid exec config type")
	}

	if attach != nil {
		attach.Stdin = execConfig.AttachStdin
	}

	io, err := cm.openExecIO(execid, attach)
	if err != nil {
		return err
	}

	process := &specs.Process{
		Args:     execConfig.Cmd,
		Terminal: execConfig.Tty,
		Cwd:      "/",
	}

	return cm.Client.ExecContainer(ctx, &ctrd.Process{
		ContainerID: execConfig.ContainerID,
		ExecID:      execid,
		IO:          io,
		P:           process,
	})
}

// Create checks passed in parameters and create a Container object whose status is set at Created.
func (cm *ContainerManager) Create(ctx context.Context, name string, config *types.ContainerConfigWrapper) (*types.ContainerCreateResp, error) {
	var id string
	for {
		id = randomid.Generate()
		_, err := cm.Store.Get(id)
		if err != nil {
			if merr, ok := err.(meta.Error); ok && merr.IsNotfound() {
				break
			}
			return nil, err
		}
	}
	if name == "" {
		i := 0
		for {
			if i+6 > len(id) {
				break
			}
			name = id[i : i+6]
			i++
			if !cm.NameToID.Get(name).Exist() {
				break
			}
		}
	}
	if cm.NameToID.Get(name).Exist() {
		return nil, fmt.Errorf("container with name %s already exist", name)
	}

	cm.km.Lock(id)
	defer cm.km.Unlock(id)

	// parse volume config
	if err := cm.parseVolumes(ctx, config); err != nil {
		return nil, errors.Wrap(err, "failed to parse volume argument")
	}

	// TODO add more validation of parameter
	// TODO check whether image exist
	c := &types.ContainerInfo{
		ContainerState: &types.ContainerState{
			Status: types.CREATED,
		},
		ID:     id,
		Name:   name,
		Config: config,
	}

	//add to collection
	cm.NameToID.Put(name, id)
	cm.Store.Put(c)
	return &types.ContainerCreateResp{
		ID:   id,
		Name: name,
	}, nil
}

// Start a pre created Container.
func (cm *ContainerManager) Start(ctx context.Context, cfg types.ContainerStartConfig) (err error) {
	if cfg.ID == "" {
		return fmt.Errorf("either container name or id is required")
	}

	cm.km.Lock(cfg.ID)
	defer cm.km.Unlock(cfg.ID)

	c, err := cm.containerInfo(cfg.ID)
	if err != nil {
		return err
	}
	if c == nil || c.Config == nil || c.ContainerState == nil {
		return fmt.Errorf("no container found by %s", cfg.ID)
	}
	c.DetachKeys = cfg.DetachKeys

	// new a default spec.
	s, err := ctrd.NewDefaultSpec(ctx, c.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to generate spec: %s", c.ID)
	}

	for _, f := range spec.SetupFuncs() {
		if err = f(ctx, c, s); err != nil {
			return err
		}
	}

	// open container's stdio.
	io, err := cm.openContainerIO(c.ID, nil)
	if err != nil {
		return errors.Wrap(err, "failed to open io")
	}
	if io.Stdin.OpenStdin() {
		s.Process.Terminal = true
	}

	err = cm.Client.CreateContainer(ctx, &ctrd.Container{
		Info: c,
		Spec: s,
		IO:   io,
	})
	if err == nil {
		c.Status = types.RUNNING
		c.StartedAt = time.Now()
		//TODO get and set container pid
	} else {
		c.FinishedAt = time.Now()
		c.ErrorMsg = err.Error()
		c.Pid = 0
		//TODO get and set exit code

		// release io
		io.Stderr.Close()
		io.Stdout.Close()
		io.Stdin.Close()
		cm.IOs.Remove(c.ID)
	}
	cm.Store.Put(c)
	return err
}

// Stop stops a running container.
func (cm *ContainerManager) Stop(ctx context.Context, name string, timeout time.Duration) error {
	var (
		ci  *types.ContainerInfo
		err error
	)

	if ci, err = cm.containerInfo(name); err != nil {
		return errors.Wrap(err, "failed to stop container")
	}

	cm.km.Lock(ci.ID)
	defer cm.km.Unlock(ci.ID)

	if ci.Status != types.RUNNING {
		return fmt.Errorf("container's status is not running: %d", ci.Status)
	}

	if _, err := cm.Client.DestroyContainer(ctx, ci.ID); err != nil {
		return errors.Wrapf(err, "failed to destroy container: %s", ci.ID)
	}

	return nil
}

// Attach attachs a container's io.
func (cm *ContainerManager) Attach(ctx context.Context, name string, attach *types.AttachConfig) error {
	container, err := cm.containerInfo(name)
	if err != nil {
		return err
	}

	cm.km.Lock(container.ID)
	defer cm.km.Unlock(container.ID)

	if _, err := cm.openContainerIO(container.ID, attach); err != nil {
		return err
	}

	return nil
}

// List returns the container's list.
func (cm *ContainerManager) List(ctx context.Context) ([]*types.ContainerInfo, error) {
	cis := []*types.ContainerInfo{}

	list, err := cm.Store.List()
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

// containerInfo returns the 'ContainerInfo' object, the parameter 's' may be container's
// name, id or prefix id.
func (cm *ContainerManager) containerInfo(s string) (*types.ContainerInfo, error) {
	var (
		obj meta.Object
		err error
	)

	// name is the container's name.
	id, ok := cm.NameToID.Get(s).String()
	if ok {
		if obj, err = cm.Store.Get(id); err != nil {
			return nil, errors.Wrapf(err, "failed to get container info: %s", s)
		}
	} else {
		// name is the container's prefix of the id.
		objs, err := cm.Store.GetWithPrefix(s)
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

func (cm *ContainerManager) openContainerIO(id string, attach *types.AttachConfig) (*containerio.IO, error) {
	return cm.openIO(id, attach, false)
}

func (cm *ContainerManager) openExecIO(id string, attach *types.AttachConfig) (*containerio.IO, error) {
	return cm.openIO(id, attach, true)
}

func (cm *ContainerManager) openIO(id string, attach *types.AttachConfig, exec bool) (*containerio.IO, error) {
	if io := cm.IOs.Get(id); io != nil {
		return io, nil
	}

	options := []func(*containerio.Option){
		containerio.WithID(id),
	}

	if !exec {
		root := cm.Store.Path(id)
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

	cm.IOs.Put(id, io)

	return io, nil
}

func (cm *ContainerManager) stoppedAndRelease(id string, m *ctrd.Message) error {
	cm.km.Lock(id)
	defer cm.km.Unlock(id)

	// update container info
	c, err := cm.containerInfo(id)
	if err != nil {
		return err
	}
	c.Pid = -1
	c.ExitCodeValue = int(m.ExitCode())
	c.FinishedAt = time.Now()
	c.Status = types.STOPPED

	if m.HasError() {
		c.ErrorMsg = m.Error().Error()
	}

	// release resource
	if io := cm.IOs.Get(id); io != nil {
		io.Stderr.Close()
		io.Stdout.Close()
		io.Stdin.Close()
		cm.IOs.Remove(id)
	}

	// update meta
	if err := cm.Store.Put(c); err != nil {
		logrus.Errorf("failed to update meta: %v", err)
	}
	return nil
}

func (cm *ContainerManager) exitedAndRelease(id string, m *ctrd.Message) error {
	// release io.
	if io := cm.IOs.Get(id); io != nil {
		io.Stderr.Close()
		io.Stdout.Close()
		io.Stdin.Close()
		cm.IOs.Remove(id)
	}
	cm.ExecProcesses.Remove(id)
	return nil
}

func (cm *ContainerManager) parseVolumes(ctx context.Context, c *types.ContainerConfigWrapper) error {
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
			_, err := cm.VolumeMgr.Info(ctx, source)
			if err != nil {
				opts := map[string]string{
					"backend": "local",
					"size":    "100G",
				}
				if err := cm.VolumeMgr.Create(ctx, source, c.HostConfig.VolumeDriver, opts, nil); err != nil {
					logrus.Errorf("failed to create volume: %s, err: %v", source, err)
					return errors.Wrap(err, "failed to create volume")
				}
			}

			if _, err := cm.VolumeMgr.Attach(ctx, source, nil); err != nil {
				logrus.Errorf("failed to attach volume: %s, err: %v", source, err)
				return errors.Wrap(err, "failed to attach volume")
			}

			mountPath, err := cm.VolumeMgr.Path(ctx, source)
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
