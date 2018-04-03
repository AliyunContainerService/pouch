package mgr

import (
	"context"
	"strconv"

	"github.com/sirupsen/logrus"
)

// setupAnnotations extracts other related options from HostConfig and locate them in spec's annotations which will be dealt by vendored runc.
func setupAnnotations(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s

	r := meta.HostConfig.Resources

	s.Annotations = make(map[string]string)

	if r.MemoryWmarkRatio != nil {
		s.Annotations["__memory_wmark_ratio"] = strconv.FormatInt(*r.MemoryWmarkRatio, 10)
	}

	if r.MemoryExtra != nil {
		s.Annotations["__memory_extra_in_bytes"] = strconv.FormatInt(*r.MemoryExtra, 10)
	}

	s.Annotations["__memory_force_empty_ctl"] = strconv.FormatInt(r.MemoryForceEmptyCtl, 10)

	s.Annotations["__schedule_latency_switch"] = strconv.FormatInt(r.ScheLatSwitch, 10)

	// add additional spec annotations
	annotations := meta.Config.SpecAnnotation
	for k, v := range annotations {
		if _, exist := s.Annotations[k]; exist {
			logrus.Warnf("Duplicate spec annotation: %s=%s", k, v)
			continue
		}
		s.Annotations[k] = v
	}

	return nil
}
