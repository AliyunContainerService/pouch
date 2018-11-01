// +build linux,seccomp

package mgr

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/containerd/containerd/contrib/seccomp"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// IsSeccompEnable return true since pouch support seccomp in build
func IsSeccompEnable() bool {
	return true
}

// setupSeccomp creates seccomp security settings spec.
func setupSeccomp(ctx context.Context, c *Container, s *specs.Spec) error {
	if c.HostConfig.Privileged {
		return nil
	}

	if s.Linux.Seccomp == nil {
		s.Linux.Seccomp = &specs.LinuxSeccomp{}
	}

	seccompProfile := c.SeccompProfile
	switch seccompProfile {
	case ProfileNameUnconfined:
		return nil
	case ProfilePouchDefault, "":
		s.Linux.Seccomp = seccomp.DefaultProfile(s)
	default:
		data, err := ioutil.ReadFile(seccompProfile)
		if err != nil {
			return fmt.Errorf("failed to load seccomp profile %q: %v", seccompProfile, err)
		}
		err = json.Unmarshal(data, s.Linux.Seccomp)
		if err != nil {
			return fmt.Errorf("failed to decode seccomp profile %q: %v", seccompProfile, err)
		}
	}

	return nil
}
