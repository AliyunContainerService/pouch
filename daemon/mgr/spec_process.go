package mgr

import (
	"context"
	"os"

	"github.com/alibaba/pouch/pkg/user"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
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
	// container rootfs is created by containerd, pouch just creates a snapshot
	// id and keeps it in memory. If container is in start process, we can not
	// find if user if exist in container image, so we do some simple check.
	var uid, gid uint32

	if c.Config.User != "" {
		if _, err := os.Stat(c.BaseFS); err != nil {
			logrus.Infof("snapshot %s is not exist, maybe in start process.", c.BaseFS)
			uid, gid = user.GetIntegerID(c.Config.User)
		} else {
			uid, gid, err = user.Get(c.BaseFS, c.Config.User)
			if err != nil {
				return err
			}
		}
	}

	additionalGids := user.GetAdditionalGids(c.HostConfig.GroupAdd)

	spec.s.Process.User = specs.User{
		UID:            uid,
		GID:            gid,
		AdditionalGids: additionalGids,
	}
	return nil
}

func setupNoNewPrivileges(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	if meta.HostConfig.Privileged {
		return nil
	}

	spec.s.Process.NoNewPrivileges = meta.NoNewPrivileges
	return nil
}

func setupOOMScoreAdj(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) (err error) {
	v := int(c.HostConfig.OomScoreAdj)
	spec.s.Process.OOMScoreAdj = &v
	return nil
}
