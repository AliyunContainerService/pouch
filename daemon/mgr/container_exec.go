package mgr

import (
	"context"
	"fmt"
	"io"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/docker/docker/pkg/stdcopy"
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
func (mgr *ContainerManager) StartExec(ctx context.Context, execid string, attach *AttachConfig) (err error) {
	// GetExecConfig should not error, since we have done this before call StartExec
	execConfig, err := mgr.GetExecConfig(ctx, execid)
	if err != nil {
		return err
	}

	eio, err := mgr.openExecIO(execid, attach)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			var stdout io.Writer = eio.Stdout
			if !execConfig.Tty && !eio.MuxDisabled {
				stdout = stdcopy.NewStdWriter(stdout, stdcopy.Stdout)
			}
			stdout.Write([]byte(err.Error() + "\r\n"))
			// close io to make hijack connection exit
			eio.Close()
			mgr.IOs.Remove(execid)
			// set exec exit status
			execConfig.Running = false
			exitCode := 126
			execConfig.ExitCode = int64(exitCode)
		}
		mgr.ExecProcesses.Put(execid, execConfig)
	}()

	if attach != nil {
		attach.Stdin = execConfig.AttachStdin
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

	err = mgr.Client.ExecContainer(ctx, &ctrd.Process{
		ContainerID: execConfig.ContainerID,
		ExecID:      execid,
		IO:          eio,
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

	entrypoint, args := mgr.getEntrypointAndArgs(execConfig.Cmd)
	processConfig := &types.ProcessConfig{
		Privileged: execConfig.Privileged,
		Tty:        execConfig.Tty,
		User:       execConfig.User,
		Arguments:  args,
		Entrypoint: entrypoint,
	}

	return &types.ContainerExecInspect{
		ID: execConfig.ExecID,
		// FIXME: try to use the correct running status of exec
		Running:       execConfig.Running,
		ExitCode:      execConfig.ExitCode,
		ContainerID:   execConfig.ContainerID,
		ProcessConfig: processConfig,
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

// CheckExecExist check if exec process `name` exist
func (mgr *ContainerManager) CheckExecExist(ctx context.Context, name string) error {
	_, err := mgr.GetExecConfig(ctx, name)
	return err
}

func (mgr *ContainerManager) getEntrypointAndArgs(cmd []string) (string, []string) {
	if len(cmd) == 0 {
		return "", []string{}
	}

	return cmd[0], cmd[1:]
}
