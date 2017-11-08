package spec

import (
	"context"

	"github.com/alibaba/pouch/apis/types"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupNetwork(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error {
	s.Hostname = c.Config.Hostname
	//TODO setup network parameters
	return nil
}
