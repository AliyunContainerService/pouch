package mgr

import (
	"context"
	"strconv"
)

// setupAliOsOption extracts alios related options from HostConfig and locate them in spec's annotations which will be dealt by vendored runc.
func setupAliOsOption(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s

	r := meta.HostConfig.Resources

	s.Annotations = make(map[string]string)

	if r.MemoryWmarkRatio != nil {
		s.Annotations["_MEMORY_WATER_MARK_RATIO"] = strconv.FormatInt(*r.MemoryWmarkRatio, 10)
	}

	if r.MemoryExtra != nil {
		s.Annotations["_MEMORY_EXTRA"] = strconv.FormatInt(*r.MemoryExtra, 10)
	}

	s.Annotations["_MEMORY_FORCE_EMPTY_CTL"] = strconv.FormatInt(r.MemoryForceEmptyCtl, 10)

	s.Annotations["_SCHEDULE_LATENCY_SWITCH"] = strconv.FormatInt(r.ScheLatSwitch, 10)

	return nil
}
