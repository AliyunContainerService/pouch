package mgr

import (
	"context"
)

func setupRoot(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	spec.s.Root.Readonly = meta.HostConfig.ReadonlyRootfs

	return nil
}
