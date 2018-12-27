package v1alpha2

import (
	"context"
	"fmt"
	"time"

	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
	"github.com/alibaba/pouch/cri/metrics"
	util_metrics "github.com/alibaba/pouch/pkg/utils/metrics"
)

// UnpauseContainer unpauses the container.
func (c *CriManager) UnpauseContainer(ctx context.Context, r *runtime.UnpauseContainerRequest) (*runtime.UnpauseContainerResponse, error) {
	label := util_metrics.ActionUnpauseLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	containerID := r.GetContainerId()

	if err := c.ContainerMgr.Unpause(ctx, containerID); err != nil {
		return nil, fmt.Errorf("failed to unpause container %q: %v", containerID, err)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.UnpauseContainerResponse{}, nil
}
