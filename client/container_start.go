package client

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
)

// ContainerStart starts a created container.
func (client *APIClient) ContainerStart(ctx context.Context, name string, options types.ContainerStartOptions) error {
	resp, err := client.post(ctx, "/containers/"+name+"/start", nil, &options, nil)
	ensureCloseReader(resp)

	return err
}
