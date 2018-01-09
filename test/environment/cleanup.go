package environment

import (
	"fmt"

	"github.com/alibaba/pouch/client"
	"github.com/pkg/errors"
)

// PruneAllImages deletes all images from pouchd.
func PruneAllImages(apiClient client.ImageAPIClient) error {
	images, err := apiClient.ImageList()
	if err != nil {
		return errors.Wrap(err, "fail to list images")
	}

	for _, img := range images {
		// force to remove the image
		if err := apiClient.ImageRemove(img.Name, true); err != nil {
			return errors.Wrap(err, fmt.Sprintf("fail to remove image (%s)", img.Name))
		}
	}
	return nil
}

// PruneAllContainers deletes all containers from pouchd.
func PruneAllContainers(apiClient client.ContainerAPIClient) error {
	containers, err := apiClient.ContainerList(true)
	if err != nil {
		return errors.Wrap(err, "fail to list images")
	}

	for _, ctr := range containers {
		// force to remove the containers
		if err := apiClient.ContainerRemove(ctr.ID, true); err != nil {
			return errors.Wrap(err, fmt.Sprintf("fail to remove container (%s)", ctr.ID))
		}
	}
	return nil
}
