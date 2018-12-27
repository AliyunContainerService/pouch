package v1alpha2

import (
	"context"
	"fmt"
	"time"

	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
	"github.com/alibaba/pouch/cri/metrics"
	util_metrics "github.com/alibaba/pouch/pkg/utils/metrics"
)

// PauseContainer pauses the container.
func (c *CriManager) PauseContainer(ctx context.Context, r *runtime.PauseContainerRequest) (*runtime.PauseContainerResponse, error) {
	label := util_metrics.ActionPauseLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	containerID := r.GetContainerId()

	if err := c.ContainerMgr.Pause(ctx, containerID); err != nil {
		return nil, fmt.Errorf("failed to pause container %q: %v", containerID, err)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.PauseContainerResponse{}, nil
}
