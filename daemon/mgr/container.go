package mgr

import (
	"context"
	"fmt"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
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
}

// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
	Store    *meta.Store
	Client   *ctrd.Client
	NameToID *collect.SafeMap
	ImageMgr ImageMgr
}

// Restore containers from meta store to memory and recover those container.
func (cm *ContainerManager) Restore(ctx context.Context) error {
	fn := func(obj meta.Object) error {
		if c, ok := obj.(*types.ContainerInfo); ok {
			// map container's name to id.
			cm.NameToID.Put(c.Name, c.ID)

			// recover the running container.
			if c.Status == types.RUNNING {
				if err := cm.Client.RecoverContainer(ctx, c.ID); err != nil {
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
	s, err := ctrd.NewDefaultSpec(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to generate spec: %s", c.ID)
	}

	for _, f := range spec.SetupFuncs() {
		if err = f(ctx, c, s); err != nil {
			return err
		}
	}

	err = cm.Client.CreateContainer(ctx, c, s)
	if err == nil {
		c.Status = types.RUNNING
		c.StartedAt = time.Now()
		//TODO get and set container pid
	} else {
		c.FinishedAt = time.Now()
		c.ErrorMsg = err.Error()
		c.Pid = 0
		//TODO get and set exit code
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

	result, err := cm.Client.DestroyContainer(ctx, ci.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to destroy container: %s", ci.ID)
	}

	ci.Pid = -1
	ci.ExitCodeValue = int(result.ExitCode())
	ci.FinishedAt = result.ExitTime()
	ci.Status = types.STOPPED

	if result.HasError() {
		ci.ErrorMsg = result.Error().Error()
	}

	// update meta
	if err := cm.Store.Put(ci); err != nil {
		logrus.Errorf("failed to update meta: %v", err)
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
