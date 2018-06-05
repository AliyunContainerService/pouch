package mgr

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/randomid"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

// CreateExec creates exec process's meta data.
func (mgr *ContainerManager) CreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (string, error) {
	c, err := mgr.container(name)
	if err != nil {
		return "", err
	}

	if !c.IsRunning() {
		return "", fmt.Errorf("container %s is not running", c.ID)
	}

	execid := randomid.Generate()
	execConfig := &ContainerExecConfig{
		ExecID:           execid,
		ExecCreateConfig: *config,
		ContainerID:      c.ID,
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
	execConfig, ok := v.(*ContainerExecConfig)
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

	c.Lock()
	process := &specs.Process{
		Args:     execConfig.Cmd,
		Terminal: execConfig.Tty,
		Cwd:      "/",
		Env:      c.Config.Env,
	}

	if execConfig.User != "" {
		c.Config.User = execConfig.User
	}

	if err = setupUser(ctx, c, &specs.Spec{Process: process}); err != nil {
		c.Unlock()
		return err
	}

	// set exec process ulimit
	if err := setupRlimits(ctx, c.HostConfig, &specs.Spec{Process: process}); err != nil {
		c.Unlock()
		return err
	}
	c.Unlock()

	execConfig.Running = true
	defer func() {
		if err != nil {
			execConfig.Running = false
			exitCode := 126
			execConfig.ExitCode = int64(exitCode)
		}
		mgr.ExecProcesses.Put(execid, execConfig)
	}()

	err = mgr.Client.ExecContainer(ctx, &ctrd.Process{
		ContainerID: execConfig.ContainerID,
		ExecID:      execid,
		IO:          io,
		P:           process,
	})

	return err
}

// InspectExec returns low-level information about exec command.
func (mgr *ContainerManager) InspectExec(ctx context.Context, execid string) (*types.ContainerExecInspect, error) {
	execConfig, err := mgr.GetExecConfig(ctx, execid)
	if err != nil {
		return nil, err
	}

	return &types.ContainerExecInspect{
		ID: execConfig.ExecID,
		// FIXME: try to use the correct running status of exec
		Running:     execConfig.Running,
		ExitCode:    execConfig.ExitCode,
		ContainerID: execConfig.ContainerID,
	}, nil
}

// GetExecConfig returns execonfig of a exec process inside container.
func (mgr *ContainerManager) GetExecConfig(ctx context.Context, execid string) (*ContainerExecConfig, error) {
	v, ok := mgr.ExecProcesses.Get(execid).Result()
	if !ok {
		return nil, errors.Wrap(errtypes.ErrNotfound, "exec process "+execid)
	}
	execConfig, ok := v.(*ContainerExecConfig)
	if !ok {
		return nil, fmt.Errorf("invalid exec config type")
	}
	return execConfig, nil
}
