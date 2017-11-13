package mgr

import (
	"context"
	"fmt"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/spec"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/alibaba/pouch/pkg/randomid"

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
}

// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
	Store     *meta.Store
	Client    *ctrd.Client
	NameToID  *collect.SafeMap
	ImageMgr  ImageMgr
	VolumeMrg VolumeMgr
	IOs       *containerio.Cache
}

// NewContainerManager creates a brand new container manager.
func NewContainerManager(ctx context.Context, store *meta.Store, cli *ctrd.Client, imgMgr ImageMgr, volMgr VolumeMgr) (*ContainerManager, error) {
	mgr := &ContainerManager{
		Store:     store,
		NameToID:  collect.NewSafeMap(),
		Client:    cli,
		ImageMgr:  imgMgr,
		VolumeMrg: volMgr,
		IOs:       containerio.NewCache(),
	}

	var exitHooks []func(string, *ctrd.Message) error
	exitHooks = append(exitHooks, mgr.stoppedAndRelease)
	mgr.Client.SetExitHooks(exitHooks)

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
				}
			}
		}
		return nil
	}
	return cm.Store.ForEach(fn)
}

// Create checks passed in parameters and create a Container object whose status is set at Created.
func (cm *ContainerManager) Create(ctx context.Context, name string, config *types.ContainerConfigWrapper) (t *types.ContainerCreateResp, err error) {
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
	//TODO add more validation of parameter
	//TODO check whether image exist

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
func (cm *ContainerManager) Start(ctx context.Context, startCfg types.ContainerStartConfig) (err error) {
	if startCfg.ID == "" {
		return fmt.Errorf("either container name or id is required")
	}

	c, err := cm.containerInfo(startCfg.ID)
	if err != nil {
		return err
	}
	if c == nil || c.Config == nil || c.ContainerState == nil {
		return fmt.Errorf("no container found by %s", startCfg.ID)
	}
	c.DetachKeys = startCfg.DetachKeys

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

		// close io resource.
		if io := cm.IOs.Get(c.ID); io != nil {
			io.Stderr.Close()
			io.Stdout.Close()
			io.Stdin.Close()
			cm.IOs.Remove(c.ID)
		}
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

	if _, err := cm.openContainerIO(container.ID, attach); err != nil {
		return err
	}

	return nil
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
	if io := cm.IOs.Get(id); io != nil {
		return io, nil
	}

	root := cm.Store.Path(id)

	options := []func(*containerio.Option){
		containerio.WithID(id),
		containerio.WithRootDir(root),
	}

	if attach != nil {
		options = append(options, containerio.WithHijack(attach.Hijack, attach.Upgrade))
		if attach.Stdin {
			options = append(options, containerio.WithStdinHijack())
		}
	} else {
		options = append(options, containerio.WithRawFile())
	}

	io := containerio.NewIO(containerio.NewOption(options...))

	cm.IOs.Put(id, io)

	return io, nil
}

func (cm *ContainerManager) stoppedAndRelease(id string, m *ctrd.Message) error {
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
