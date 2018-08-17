// +build !linux

package mgr

import (
	"context"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// do nothing in this case
func setupAppArmor(ctx context.Context, c *Container, s *specs.Spec) error {
	return nil
}
