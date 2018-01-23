package mgr

import (
	"context"
)

func setupNetwork(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s

	s.Hostname = c.Config.Hostname.String()
	//TODO setup network parameters
	return nil
}
