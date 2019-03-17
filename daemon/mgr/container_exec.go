package mgr

import (
	"context"
	"fmt"
	"io"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/randomid"
	"github.com/alibaba/pouch/pkg/streams"
	"github.com/alibaba/pouch/pkg/user"
	"github.com/docker/docker/daemon/caps"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

// CreateExec creates exec process's meta data.
func (mgr *ContainerManager) CreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (string, error) {
	c, err := mgr.container(name)
	if err != nil {
		return "", err
	}

	if !c.State.Running {
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
func (mgr *ContainerManager) StartExec(ctx context.Context, execid string, cfg *streams.AttachConfig) (err0 error) {
	// GetExecConfig should not error, since we have done this before call StartExec
	execConfig, err := mgr.GetExecConfig(ctx, execid)
	if err != nil {
		return err
	}

	c, err := mgr.container(execConfig.ContainerID)
	if err != nil {
		return err
	}

	// set exec process user, user decided by exec config
	if execConfig.User == "" {
		execConfig.User = c.Config.User
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

	if execConfig.Privileged {
		capList := caps.GetAllCapabilities()
		process.Capabilities = &specs.LinuxCapabilities{
			Effective:   capList,
			Bounding:    capList,
			Permitted:   capList,
			Inheritable: capList,
		}
	} else if spec, err := mgr.getContainerSpec(c); err == nil {
		// NOTE: if container is created by docker and taken over by pouchd,
		// no config.json can found under current path, runc exec is good even
		// without these capabilities in exec process
		process.Capabilities = spec.Process.Capabilities
	}

	// set exec process ulimit, ulimit not decided by exec config
	if err := setupRlimits(ctx, c.HostConfig, &specs.Spec{Process: process}); err != nil {
		return err
	}

	// NOTE: the StartExec might use the hijack's connection as
	// stdin in the AttachConfig. If we close it directly, the stdout/stderr
	// will return the `using closed connection` error. As a result, the
	// Attach will return the error. We need to use pipe here instead of
	// origin one and let the caller closes the stdin by themself.
	if execConfig.AttachStdin && cfg.UseStdin {
		oldStdin := cfg.Stdin
		pstdinr, pstdinw := io.Pipe()
		go func() {
			defer pstdinw.Close()
			io.Copy(pstdinw, oldStdin)
		}()
		cfg.Stdin = pstdinr
	} else {
		cfg.UseStdin = false
	}

	// NOTE: always close stdin pipe for exec process
	cfg.CloseStdin = true
	eio, err := mgr.initExecIO(execid, cfg.UseStdin)
	if err != nil {
		return err
	}

	attachErrCh := eio.Stream().Attach(ctx, cfg)

	defer func() {
		if err0 != nil {
			// set exec exit status
			execConfig.Running = false
			exitCode := 126
			execConfig.ExitCode = int64(exitCode)
			eio.Close()
			mgr.IOs.Remove(execid)
		}
	}()

	execConfig.Running = true
	if err := mgr.Client.ExecContainer(ctx, &ctrd.Process{
		ContainerID: execConfig.ContainerID,
		ExecID:      execid,
		IO:          eio,
		P:           process,
	}); err != nil {
		return err
	}
	return <-attachErrCh
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
