// +build linux,!seccomp

package mgr

import (
	"context"
	"fmt"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// IsSeccompEnable return false since pouch do not support seccomp in build
func IsSeccompEnable() bool {
	return false
}

func setupSeccomp(ctx context.Context, c *Container, s *specs.Spec) error {
	if c.SeccompProfile != "" && c.SeccompProfile != "unconfined" {
		return fmt.Errorf("Seccomp is not support by pouch, can not set seccomp profile %s", c.SeccompProfile)
	}

	return nil
}
