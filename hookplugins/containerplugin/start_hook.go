package containerplugin

import (
	"context"

	networktypes "github.com/alibaba/pouch/network/types"
)

// PreStart returns an array of priority and args which will pass to runc, the every priority
// used to sort the pre start array that pass to runc, network plugin hook always has priority value 0.
func (c *contPlugin) PreStart(ctx context.Context, config interface{}) ([]int, [][]string, error) {
	// TODO: Implemented by the developer
	return nil, nil, nil
}

// PreCreateEndpoint accepts the container id and env of this container, to update the config of container's endpoint.
func (c *contPlugin) PreCreateEndpoint(ctx context.Context, cid string, env []string, endpoint *networktypes.Endpoint) error {
	// TODO: Implemented by the developer
	return nil
}
