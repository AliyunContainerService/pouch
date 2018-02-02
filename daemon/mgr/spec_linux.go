package mgr

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/docker/docker/daemon/caps"
)

const (
	// ProfileNamePrefix is the prefix for loading profiles on a localhost. Eg. localhost/profileName.
	ProfileNamePrefix = "localhost/"
	// ProfileRuntimeDefault indicates that we should use or create a runtime default profile.
	ProfileRuntimeDefault = "runtime/default"
	// ProfileNameUnconfined is a string indicating one should run a pod/containerd without a security profile.
	ProfileNameUnconfined = "unconfined"
)

// Setup linux-platform-sepecific specification.

func setupSysctl(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	spec.s.Linux.Sysctl = meta.HostConfig.Sysctls
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

func setupAppArmor(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	if !isAppArmorEnabled() {
		// Return if the apparmor is disabled.
		return nil
	}

	appArmorProfile := meta.AppArmorProfile
	switch appArmorProfile {
	case ProfileNameUnconfined:
		return nil
	case ProfileRuntimeDefault, "":
		// TODO: handle runtime default case.
		return nil
	default:
		spec.s.Process.ApparmorProfile = appArmorProfile
	}

	return nil
}

func setupCapabilities(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	var caplist []string
	var err error

	capabilities := spec.s.Process.Capabilities
	if meta.HostConfig.Privileged {
		caplist = caps.GetAllCapabilities()
	} else if caplist, err = caps.TweakCapabilities(capabilities.Effective, meta.HostConfig.CapAdd, meta.HostConfig.CapDrop); err != nil {
		return err
	}
	capabilities.Effective = caplist
	capabilities.Bounding = caplist
	capabilities.Permitted = caplist
	capabilities.Inheritable = caplist

	return nil
}
