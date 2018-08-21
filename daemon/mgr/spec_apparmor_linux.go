// +build linux,apparmor

package mgr

import (
	"context"

	"github.com/opencontainers/runc/libcontainer/apparmor"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupAppArmor(ctx context.Context, c *Container, s *specs.Spec) error {
	if apparmor.IsEnabled() {
		appArmorProfile := c.AppArmorProfile

		switch appArmorProfile {
		case "runtime/default":
			return nil
		case "unconfined":
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
	}
	return nil
}
