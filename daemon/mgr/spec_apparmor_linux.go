// +build apparmor,linux

package mgr

import (
	"context"

	"github.com/opencontainers/runc/libcontainer/apparmor"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupAppArmor(ctx context.Context, c *Container, s *specs.Spec) error {
	if apparmor.IsEnabled() {
		appArmorProfile := ""
		if c.AppArmorProfile != "" {
			appArmorProfile = c.AppArmorProfile
		} else if c.HostConfig.Privileged {
			appArmorProfile = "unconfined"
		} else {
			// TODO: generate pouch-default apparmor profile
			// appArmorProfile = "pouch-default"
		}

		s.Process.ApparmorProfile = appArmorProfile
	}

	return nil
}
