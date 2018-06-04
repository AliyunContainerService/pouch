package environment

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"

	"github.com/pkg/errors"
)

// PruneAllImages deletes all images from pouchd.
func PruneAllImages(apiClient client.ImageAPIClient) error {
	ctx := context.Background()
	images, err := apiClient.ImageList(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to list images")
	}

	for _, img := range images {
		// force to remove the image
		if err := apiClient.ImageRemove(ctx, img.ID, true); err != nil {
			return errors.Wrap(err, fmt.Sprintf("fail to remove image (%s)", img.ID))
		}
	}
	return nil
}

// PruneAllContainers deletes all containers from pouchd.
func PruneAllContainers(apiClient client.ContainerAPIClient) error {
	ctx := context.Background()
	containers, err := apiClient.ContainerList(ctx, true)
	if err != nil {
		return errors.Wrap(err, "fail to list images")
	}

	for _, ctr := range containers {
		// force to remove the containers
		if err := apiClient.ContainerRemove(ctx, ctr.ID, &types.ContainerRemoveOptions{Force: true}); err != nil {
			return errors.Wrap(err, fmt.Sprintf("fail to remove container (%s)", ctr.ID))
		}
	}
	return nil
}
