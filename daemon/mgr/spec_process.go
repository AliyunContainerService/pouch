package mgr

import (
	"context"
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
	// The user option is complicated, now we only handle case "uid".
	// TODO: handle other cases like "user", "uid:gid", etc.
	if c.Config.User == "" {
		return nil
	}

	fields := strings.Split(c.Config.User, ":")
	var u string
	u = fields[0]
	user := specs.User{}
	if uid, err := strconv.Atoi(u); err == nil {
		user.UID = uint32(uid)
	} else {
		user.Username = u
	}
	spec.s.Process.User = user

	return nil
}
