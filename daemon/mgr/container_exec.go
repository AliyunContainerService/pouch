package mgr

import (
	"context"
	"fmt"
	"io"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/randomid"
	"github.com/alibaba/pouch/pkg/user"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/opencontainers/runtime-spec/specs-go"
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

	envs, err := mergeEnvSlice(config.Env, c.Config.Env)

	if err != nil {
		return "", err
	}

	execid := randomid.Generate()
	execConfig := &ContainerExecConfig{
		ExecID:           execid,
		ExecCreateConfig: *config,
		ContainerID:      c.ID,
		Env:              envs,
	}

	mgr.ExecProcesses.Put(execid, execConfig)

	return execid, nil
}

// ResizeExec resizes the size of exec process's tty.
func (mgr *ContainerManager) ResizeExec(ctx context.Context, execid string, opts types.ResizeOptions) error {
	execConfig, err := mgr.GetExecConfig(ctx, execid)
	if err != nil {
		return err
	}

	return mgr.Client.ResizeExec(ctx, execConfig.ContainerID, execid, opts)
}

// StartExec executes a new process in container.
func (mgr *ContainerManager) StartExec(ctx context.Context, execid string, attach *AttachConfig) (err error) {
	// GetExecConfig should not error, since we have done this before call StartExec
	execConfig, err := mgr.GetExecConfig(ctx, execid)
	if err != nil {
		return err
	}

	// FIXME(fuweid): make attachConfig consistent with execConfig
	if attach != nil {
		attach.Stdin = execConfig.AttachStdin
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
			// set exec exit status
			execConfig.Lock()
			execConfig.Running = false
			exitCode := 126
			execConfig.ExitCode = int64(exitCode)
			execConfig.Unlock()

			// close io to make hijack connection exit
			eio.Close()
			mgr.IOs.Remove(execid)
		}
	}()

	c, err := mgr.container(execConfig.ContainerID)
	if err != nil {
		return err
	}

	// set exec process user, user decided by exec config
	if execConfig.User == "" {
		execConfig.Lock()
		execConfig.User = c.Config.User
		execConfig.Unlock()
	}

	uid, gid, additionalGids, err := user.Get(c.GetSpecificBasePath(user.PasswdFile),
		c.GetSpecificBasePath(user.GroupFile), execConfig.User, c.HostConfig.GroupAdd)
	if err != nil {
		return err
	}

	cwd := c.Config.WorkingDir
	if cwd == "" {
		cwd = "/"
	}

	process := &specs.Process{
		Args:     execConfig.Cmd,
		Terminal: execConfig.Tty,
		Cwd:      cwd,
		Env:      execConfig.Env,
		User: specs.User{
			UID:            uid,
			GID:            gid,
			AdditionalGids: additionalGids,
		},
	}

	// set exec process ulimit, ulimit not decided by exec config
	if err := setupRlimits(ctx, c.HostConfig, &specs.Spec{Process: process}); err != nil {
		return err
	}

	execConfig.Lock()
	execConfig.Running = true
	execConfig.Unlock()

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
		return nil, errors.Wrapf(errtypes.ErrNotfound, "exec process %s", execid)
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
