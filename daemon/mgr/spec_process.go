package mgr

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupCap(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	//TODO setup capabilities
	return nil
}

func setupProcessArgs(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	args := c.Config.Entrypoint
	if args == nil {
		args = []string{}
	}
	if len(c.Config.Cmd) > 0 {
		args = append(args, c.Config.Cmd...)
	}
	s.Process.Args = args
	return nil
}

func setupProcessEnv(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	if s.Process.Env == nil {
		s.Process.Env = c.Config.Env
	} else {
		s.Process.Env = append(s.Process.Env, c.Config.Env...)
	}

	//set env for rich container mode
	s.Process.Env = append(s.Process.Env, richContainerModeEnv(c)...)

	return nil
}

func setupProcessCwd(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	if c.Config.WorkingDir == "" {
		s.Process.Cwd = "/"
	} else {
		s.Process.Cwd = c.Config.WorkingDir
	}
	return nil
}

func setupProcessTTY(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	s.Process.Terminal = c.Config.Tty
	if s.Process.Env != nil {
		s.Process.Env = append(s.Process.Env, "TERM=xterm")
	} else {
		s.Process.Env = []string{"TERM=xterm"}
	}
	return nil
}

func setupProcessUser(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) (err error) {
	if c.Config.User != "" {
		fields := strings.SplitN(c.Config.User, ":", 2)
		var u, g string
		u = fields[0]
		if len(fields) == 2 {
			g = fields[1]
		}
		user := &specs.User{}
		if uid, err := strconv.Atoi(u); err == nil {
			user.UID = uint32(uid)
		} else {
			user.Username = u
		}
		gid, err := strconv.Atoi(g)
		if err != nil || gid <= 0 {
			return fmt.Errorf("invalid gid: %d", gid)
		}
		user.GID = uint32(gid)
	}

	//TODO security config (including both seccomp and selinux)

	return nil
}
