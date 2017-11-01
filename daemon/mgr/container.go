package mgr

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/spec"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/alibaba/pouch/pkg/randomid"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

//ContainerMgr as an interface defines all operations against container.
type ContainerMgr interface {
	// Create a container.
	Create(context context.Context, name string, config *types.ContainerConfigWrapper) (*types.ContainerCreateResp, error)
	// Start a container.
	Start(context context.Context, config types.ContainerStartConfig) error
}

// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
	Store    *meta.Store
	Client   *ctrd.Client
	NameToID *collect.SafeMap
	ImageMgr ImageMgr
}

// NewContainerManager creates a brand new container manager.
func NewContainerManager(ctx context.Context, store *meta.Store) (*ContainerManager, error) {
	cm := &ContainerManager{
		Store:    store,
		NameToID: collect.NewSafeMap(),
	}

	spec.RegisteSetupFunc(setupProcess)
	spec.RegisteSetupFunc(setupNs)
	spec.RegisteSetupFunc(setupCap)

	return cm, cm.Restore(ctx)
}

// Restore containers from meta store to memory.
func (cm *ContainerManager) Restore(ctx context.Context) error {
	fn := func(obj meta.Object) error {
		if c, ok := obj.(*types.ContainerInfo); ok {
			cm.NameToID.Put(c.Name, c.ID)
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
	var obj meta.Object
	idByName, exist := cm.NameToID.Get(startCfg.ID).String()
	if exist {
		obj, err = cm.Store.Get(idByName)
		if err != nil {
			return errors.Wrapf(err, "fetch container from store error %s", idByName)
		}
	}

	if obj == nil {
		//TODO matching by prefix
		obj, err = cm.Store.Get(startCfg.ID)
		if err != nil {
			return errors.Wrapf(err, "fetch container from store error %s", startCfg.ID)
		}
	}
	c := obj.(*types.ContainerInfo)
	if c == nil || c.Config == nil || c.ContainerState == nil {
		return fmt.Errorf("no container found by %s", startCfg.ID)
	}
	c.DetachKeys = startCfg.DetachKeys

	// new a default spec.
	s, err := ctrd.NewDefaultSpec(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to generate spec: %s", c.ID)
	}

	setupSpecFuncs := spec.GetSetupFunc()
	for _, f := range setupSpecFuncs {
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

type setupFunc func(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error

func (sf setupFunc) run(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error {
	return sf(ctx, c, s)
}

// TODO: move to network module.
func setupNetwork(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error {
	s.Hostname = c.Config.Hostname
	//TODO setup network parameters
	return nil
}

func setupCap(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error {
	//TODO setup capabilities
	return nil
}

func setupNs(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error {
	//TODO setup ns
	return nil
}

func setupProcess(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) (err error) {
	p := s.Process
	cmdArr := c.Config.Entrypoint
	if len(c.Config.Cmd) > 0 {
		cmdArr = append(cmdArr, c.Config.Cmd...)
	}
	if c.Config.Tty != nil {
		p.Terminal = *c.Config.Tty
	}
	p.Cwd = c.Config.WorkingDir
	p.Env = c.Config.Env
	p.Args = cmdArr
	if c.Config.User != "" {
		tmpArr := strings.SplitN(c.Config.User, ":", 2)
		var u, g string
		u = tmpArr[0]
		if len(tmpArr) == 2 {
			g = tmpArr[1]
		}
		user := &specs.User{}
		if uid, err := strconv.Atoi(u); err == nil {
			user.UID = uint32(uid)
		} else {
			user.Username = u
		}
		if gid, err := strconv.Atoi(g); err == nil {
			user.GID = uint32(gid)
		}
	}

	//TODO security config (including both seccomp and selinux)

	return nil
}
