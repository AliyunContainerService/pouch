package mgr

import (
	"context"
	"fmt"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupMounts(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	mounts := s.Mounts
	if c.HostConfig == nil {
		return nil
	}
	for _, v := range c.HostConfig.Binds {
		sd := strings.Split(v, ":")
		lensd := len(sd)
		if lensd < 2 || lensd > 3 {
			return fmt.Errorf("unknown bind: %s", v)
		}
		opt := []string{"rbind"}
		if lensd == 3 {
			opt = append(opt, strings.Split(sd[2], ",")...)
			// Set rootfs propagation, default setting is private.
			if strings.Contains(sd[2], "rshared") {
				s.Linux.RootfsPropagation = "rshared"
			}
			if strings.Contains(sd[2], "rslave") && s.Linux.RootfsPropagation != "rshared" {
				s.Linux.RootfsPropagation = "rslave"
			}
		}
		mounts = append(mounts, specs.Mount{
			Destination: sd[1],
			Source:      sd[0],
			Type:        "bind",
			Options:     opt,
		})
	}
	s.Mounts = mounts
	return nil
}
