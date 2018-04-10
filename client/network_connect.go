package client

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
)

// NetworkConnect connects a container to a network.
func (client *APIClient) NetworkConnect(ctx context.Context, req *types.NetworkConnectConfig) error {
	resp, err := client.post(ctx, "/networks/{id:.*}/connect", nil, req, nil)
	if err != nil {
		return err
	}

	ensureCloseReader(resp)
	return nil
}
