package mgr

import (
	"context"

	"github.com/alibaba/pouch/apis/types"

	"github.com/sirupsen/logrus"
)

// Prune implement interface
func (mgr *ContainerManager) Prune(ctx context.Context, option *ContainerListOption) (*types.ContainerPruneResp, error) {
	var rep types.ContainerPruneResp

	containerLists, err := mgr.List(ctx, option)

	if err != nil {
		return nil, err
	}

	for _, c := range containerLists {
		if c.IsStopped() {
			if c.SizeRw > 0 {
				rep.SpaceReclaimed += c.SizeRw
			}

			err := mgr.Store.Remove(c.ID)
			if err != nil {
				logrus.Warnf("failed to prune container %s: %v", c.ID, err)
				continue
			}

			rep.ContainersDeleted = append(rep.ContainersDeleted, c.ID)
		}
	}

	return &rep, nil
}
