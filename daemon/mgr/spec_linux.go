package mgr

import (
	"context"
)

// Setup linux-platform-sepecific specification.

func setupSysctl(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	spec.s.Linux.Sysctl = meta.HostConfig.Sysctls
	return nil
}
