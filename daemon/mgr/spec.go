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

// All the functions related to the spec is lock-free for container instance,
// so when calling functions here like createSpec, setupProcess, setupMounts,
// setupUser and so on, caller should explicitly add lock for container instance.

// createSpec create a runtime-spec.
func createSpec(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	c.Lock()
	defer c.Unlock()
	// new a default spec from containerd.
	s, err := ctrd.NewDefaultSpec(ctx, c.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to generate spec: %s", c.ID)
	}

	// fix https://github.com/opencontainers/runc/issues/705
	// use ssh connect to container, can't use sudo command.
	s.Process.NoNewPrivileges = false

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
	return populatePlatform(ctx, c, specWrapper)
}
