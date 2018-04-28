package mgr

import (
	"context"

	"github.com/alibaba/pouch/ctrd"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

// SpecWrapper wraps the container's specs and add manager operations.
type SpecWrapper struct {
	s *specs.Spec

	ctrMgr  ContainerMgr
	volMgr  VolumeMgr
	netMgr  NetworkMgr
	prioArr []int
	argsArr [][]string
}

// createSpec create a runtime-spec.
func createSpec(ctx context.Context, c *ContainerMeta, specWrapper *SpecWrapper) error {
	// new a default spec from containerd.
	s, err := ctrd.NewDefaultSpec(ctx, c.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to generate spec: %s", c.ID)
	}
	specWrapper.s = s

	s.Hostname = c.Config.Hostname.String()
	s.Root = &specs.Root{
		Path:     c.BaseFS,
		Readonly: c.HostConfig.ReadonlyRootfs,
	}

	// create Spec.Process spec
	if err := setupProcess(ctx, c, s); err != nil {
		return err
	}

	// create Spec.Mounts spec
	if err := setupMounts(ctx, c, s); err != nil {
		return err
	}

	// create Spec.Annotations
	if err := setupAnnotations(ctx, c, s); err != nil {
		return err
	}

	// create Spec.Hooks spec
	if err := setupHook(ctx, c, specWrapper); err != nil {
		return err
	}

	// platform-specified spec setting
	// TODO: support window and Solaris platform
	if err := populatePlatform(ctx, c, specWrapper); err != nil {
		return err
	}

	return nil
}
