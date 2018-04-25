package mgr

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/user"

	"github.com/docker/docker/daemon/caps"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

// setupProcess setups spec process.
func setupProcess(ctx context.Context, c *ContainerMeta, s *specs.Spec) error {
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

	if err := setupAppArmor(ctx, c, s); err != nil {
		return err
	}

	return nil
}

func createEnvironment(c *ContainerMeta) []string {
	env := c.Config.Env
	env = append(env, richContainerModeEnv(c)...)

	return env
}

func setupUser(ctx context.Context, c *ContainerMeta, s *specs.Spec) (err error) {
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

// isAppArmorEnabled returns true if apparmor is enabled for the host.
// This function is forked from
// https://github.com/opencontainers/runc/blob/1a81e9ab1f138c091fe5c86d0883f87716088527/libcontainer/apparmor/apparmor.go
// to avoid the libapparmor dependency.
func isAppArmorEnabled() bool {
	if _, err := os.Stat("/sys/kernel/security/apparmor"); err == nil && os.Getenv("container") == "" {
		if _, err = os.Stat("/sbin/apparmor_parser"); err == nil {
			buf, err := ioutil.ReadFile("/sys/module/apparmor/parameters/enabled")
			return err == nil && len(buf) > 1 && buf[0] == 'Y'
		}
	}
	return false
}

func setupAppArmor(ctx context.Context, c *ContainerMeta, s *specs.Spec) error {
	if !isAppArmorEnabled() {
		// Return if the apparmor is disabled.
		return nil
	}

	appArmorProfile := c.AppArmorProfile
	switch appArmorProfile {
	case ProfileNameUnconfined:
		return nil
	case ProfileRuntimeDefault:
		// TODO: handle runtime default case.
		return nil
	case "":
		if c.HostConfig.Privileged {
			return nil
		}
		// TODO: if user does not specify the AppArmor and the container is not in privilege mode,
		// we need to specify it as default case, handle it later.
		return nil
	default:
		s.Process.ApparmorProfile = appArmorProfile
	}

	return nil
}
