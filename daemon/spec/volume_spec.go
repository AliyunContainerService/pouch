package spec

import (
	"context"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupMounts(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error {
	mounts := s.Mounts
	if c.HostConfig == nil {
		return nil
	}
	for _, v := range c.HostConfig.Binds {
		sd := strings.SplitN(v, ":", 2)
		if len(sd) != 2 {
			return fmt.Errorf("unknown bind: %s", v)
		}
		mounts = append(mounts, specs.Mount{
			Destination: sd[0],
			Source:      sd[1],
			Type:        "bind",
			Options:     []string{"rbind"},
		})
	}
	s.Mounts = mounts
	return nil
}
