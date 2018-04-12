package mgr

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/containerd/containerd/contrib/seccomp"
	"github.com/docker/docker/daemon/caps"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const (
	// ProfileNamePrefix is the prefix for loading profiles on a localhost. Eg. localhost/profileName.
	ProfileNamePrefix = "localhost/"
	// ProfileRuntimeDefault indicates that we should use or create a runtime default profile.
	ProfileRuntimeDefault = "runtime/default"
	// ProfileDockerDefault indicates that we should use or create a docker default profile.
	ProfileDockerDefault = "docker/default"
	// ProfilePouchDefault indicates that we should use or create a pouch default profile.
	ProfilePouchDefault = "pouch/default"
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
	case ProfileRuntimeDefault:
		// TODO: handle runtime default case.
		return nil
	case "":
		if meta.HostConfig.Privileged {
			return nil
		}
		// TODO: if user does not specify the AppArmor and the container is not in privilege mode,
		// we need to specify it as default case, handle it later.
		return nil
	default:
		spec.s.Process.ApparmorProfile = appArmorProfile
	}

	return nil
}

func setupSeccomp(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	if meta.HostConfig.Privileged {
		return nil
	}

	// TODO: check whether seccomp is enable in your kernel, if not, cannot run a custom seccomp prifle.
	seccompProfile := meta.SeccompProfile
	switch seccompProfile {
	case ProfileNameUnconfined:
		return nil
	case ProfilePouchDefault, "":
		spec.s.Linux.Seccomp = seccomp.DefaultProfile(spec.s)
	default:
		spec.s.Linux.Seccomp = &specs.LinuxSeccomp{}
		data, err := ioutil.ReadFile(seccompProfile)
		if err != nil {
			return fmt.Errorf("failed to load seccomp profile %q: %v", seccompProfile, err)
		}
		err = json.Unmarshal(data, spec.s.Linux.Seccomp)
		if err != nil {
			return fmt.Errorf("failed to decode seccomp profile %q: %v", seccompProfile, err)
		}
	}

	return nil
}

func setupSELinux(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	if !meta.HostConfig.Privileged {
		spec.s.Process.SelinuxLabel = meta.ProcessLabel
		spec.s.Linux.MountLabel = meta.MountLabel
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

func setupIntelRdt(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s

	if meta.HostConfig.IntelRdtL3Cbm != "" {
		s.Linux.IntelRdt = &specs.LinuxIntelRdt{
			L3CacheSchema: meta.HostConfig.IntelRdtL3Cbm,
		}
	}

	return nil
}
