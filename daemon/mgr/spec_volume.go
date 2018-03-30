package mgr

import (
	"context"
	"fmt"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupMounts(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	mounts := s.Mounts
	if c.HostConfig == nil {
		return nil
	}
	for _, mp := range c.Mounts {
		// check duplicate mountpoint
		for _, sm := range mounts {
			if sm.Destination == mp.Destination {
				return fmt.Errorf("duplicate mount point: %s", mp.Destination)
			}
		}

		pg := mp.Propagation
		rootfspg := s.Linux.RootfsPropagation
		// Set rootfs propagation, default setting is private.
		switch pg {
		case "shared", "rshared":
			if rootfspg != "shared" && rootfspg != "rshared" {
				s.Linux.RootfsPropagation = "shared"
			}
		case "slave", "rslave":
			if rootfspg != "shared" && rootfspg != "rshared" && rootfspg != "slave" && rootfspg != "rslave" {
				s.Linux.RootfsPropagation = "rslave"
			}
		}

		opts := []string{"rbind"}
		if !mp.RW {
			opts = append(opts, "ro")
		}
		if pg != "" {
			opts = append(opts, pg)
		}

		// TODO: support copy data.

		mounts = append(mounts, specs.Mount{
			Source:      mp.Source,
			Destination: mp.Destination,
			Type:        "bind",
			Options:     opts,
		})
	}
	s.Mounts = mounts
	return nil
}
