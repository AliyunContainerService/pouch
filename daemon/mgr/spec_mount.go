package mgr

import (
	"context"
	"fmt"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func clearReadonly(m *specs.Mount) {
	var opts []string
	for _, o := range m.Options {
		if o != "ro" {
			opts = append(opts, o)
		}
	}
	m.Options = opts
}

// setupMounts create mount spec.
func setupMounts(ctx context.Context, c *ContainerMeta, s *specs.Spec) error {
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

	if c.HostConfig.Privileged {
		if !s.Root.Readonly {
			// Clear readonly for /sys.
			for i := range s.Mounts {
				if s.Mounts[i].Destination == "/sys" {
					clearReadonly(&s.Mounts[i])
				}
			}
		}
	}
	return nil
}
