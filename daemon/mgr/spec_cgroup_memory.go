package mgr

import (
	"context"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func getCgroupMemory(s *specs.Spec) *specs.LinuxMemory {
	if s.Linux.Resources.Memory == nil {
		s.Linux.Resources.Memory = &specs.LinuxMemory{}
	}
	return s.Linux.Resources.Memory
}

func setupCgroupMemory(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	mem := getCgroupMemory(s)

	v := meta.HostConfig.Memory
	mem.Limit = &v
	return nil
}

func setupCgroupMemorySwap(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	mem := getCgroupMemory(s)

	v := meta.HostConfig.MemorySwap
	mem.Swap = &v
	return nil
}

func setupCgroupMemorySwappiness(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	mem := getCgroupMemory(s)

	var v uint64
	if meta.HostConfig.MemorySwappiness != nil {
		v = uint64(*(meta.HostConfig.MemorySwappiness))
	}
	mem.Swappiness = &v
	return nil
}
