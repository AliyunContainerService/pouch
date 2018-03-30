package client

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
)

// NetworkList lists all the networks.
func (client *APIClient) NetworkList(ctx context.Context) (*types.NetworkListResp, error) {
	resp, err := client.get(ctx, "/networks", nil, nil)
	if err != nil {
		return nil, err
	}

	network := &types.NetworkListResp{}

	err = decodeBody(network, resp.Body)
	ensureCloseReader(resp)

	return network, err
}
