package mgr

import (
	"context"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupNetwork(ctx context.Context, c *ContainerMeta, s *specs.Spec) error {
	s.Hostname = c.Config.Hostname
	//TODO setup network parameters
	return nil
}
