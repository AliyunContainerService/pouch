package mgr

import (
	"context"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/user"

	"github.com/docker/docker/daemon/caps"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// setupProcess setups spec process.
func setupProcess(ctx context.Context, c *Container, s *specs.Spec) error {
	if s.Process == nil {
		s.Process = &specs.Process{}
	}
	config := c.Config

	cwd := config.WorkingDir
	if cwd == "" {
		cwd = "/"
	}

	s.Process.Args = append(config.Entrypoint, config.Cmd...)
	s.Process.Env = append(s.Process.Env, createEnvironment(c)...)
	s.Process.Cwd = cwd
	s.Process.Terminal = config.Tty

	if s.Process.Terminal {
		s.Process.Env = append(s.Process.Env, "TERM=xterm")
	}

	if !c.HostConfig.Privileged {
		s.Process.SelinuxLabel = c.ProcessLabel
		s.Process.NoNewPrivileges = c.NoNewPrivileges
	}

	if err := setupUser(ctx, c, s); err != nil {
		return err
	}

	if c.HostConfig.OomScoreAdj != 0 {
		v := int(c.HostConfig.OomScoreAdj)
		s.Process.OOMScoreAdj = &v
	}

	if err := setupCapabilities(ctx, c.HostConfig, s); err != nil {
		return err
	}

	if err := setupRlimits(ctx, c.HostConfig, s); err != nil {
		return err
	}

	if err := setupAppArmor(ctx, c, s); err != nil {
		return err
	}

	return setupNvidiaEnv(ctx, c, s)
}

func createEnvironment(c *Container) []string {
	env := c.Config.Env
	env = append(env, richContainerModeEnv(c)...)

	return env
}

func setupUser(ctx context.Context, c *Container, s *specs.Spec) (err error) {
	uid, gid, additionalGids, err := user.Get(c.GetSpecificBasePath(user.PasswdFile),
		c.GetSpecificBasePath(user.GroupFile), c.Config.User, c.HostConfig.GroupAdd)
	if err != nil {
		return err
	}

	s.Process.User = specs.User{
		UID:            uid,
		GID:            gid,
		AdditionalGids: additionalGids,
	}
	return nil
}

func setupCapabilities(ctx context.Context, hostConfig *types.HostConfig, s *specs.Spec) error {
	var caplist []string
	var err error

	if s.Process.Capabilities == nil {
		s.Process.Capabilities = &specs.LinuxCapabilities{}
	}
	capabilities := s.Process.Capabilities

	if hostConfig.Privileged {
		caplist = caps.GetAllCapabilities()
	} else if caplist, err = caps.TweakCapabilities(capabilities.Effective, hostConfig.CapAdd, hostConfig.CapDrop); err != nil {
		return err
	}
	capabilities.Effective = caplist
	capabilities.Bounding = caplist
	capabilities.Permitted = caplist
	capabilities.Inheritable = caplist

	s.Process.Capabilities = capabilities
	return nil
}

func setupRlimits(ctx context.Context, hostConfig *types.HostConfig, s *specs.Spec) error {
	var rlimits []specs.POSIXRlimit
	for _, ul := range hostConfig.Ulimits {
		rlimits = append(rlimits, specs.POSIXRlimit{
			Type: "RLIMIT_" + strings.ToUpper(ul.Name),
			Hard: uint64(ul.Hard),
			Soft: uint64(ul.Soft),
		})
	}

	s.Process.Rlimits = rlimits
	return nil
}

func setupNvidiaEnv(ctx context.Context, c *Container, s *specs.Spec) error {
	n := c.HostConfig.NvidiaConfig
	if n == nil {
		return nil
	}
	s.Process.Env = append(s.Process.Env, fmt.Sprintf("NVIDIA_DRIVER_CAPABILITIES=%s", n.NvidiaDriverCapabilities))
	s.Process.Env = append(s.Process.Env, fmt.Sprintf("NVIDIA_VISIBLE_DEVICES=%s", n.NvidiaVisibleDevices))
	return nil
}
