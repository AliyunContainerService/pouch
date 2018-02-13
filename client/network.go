package client

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
)

// NetworkCreate creates a network.
func (client *APIClient) NetworkCreate(ctx context.Context, req *types.NetworkCreateConfig) (*types.NetworkCreateResp, error) {
	resp, err := client.post(ctx, "/networks/create", nil, req)
	if err != nil {
		return nil, err
	}

	network := &types.NetworkCreateResp{}

	err = decodeBody(network, resp.Body)
	ensureCloseReader(resp)

	return network, err
}

// NetworkRemove removes a network.
func (client *APIClient) NetworkRemove(ctx context.Context, networkID string) error {
	resp, err := client.delete(ctx, "/networks/"+networkID, nil)
	if err != nil {
		return err
	}

	ensureCloseReader(resp)
	return err
}

// NetworkInspect inspects a network.
func (client *APIClient) NetworkInspect(ctx context.Context, networkID string) (*types.NetworkInspectResp, error) {
	resp, err := client.get(ctx, "/networks/"+networkID, nil)
	if err != nil {
		return nil, err
	}

	network := &types.NetworkInspectResp{}

	err = decodeBody(network, resp.Body)
	ensureCloseReader(resp)

	return network, err
}

// NetworkList lists all the networks.
func (client *APIClient) NetworkList(ctx context.Context) (*types.NetworkListResp, error) {
	resp, err := client.get(ctx, "/networks", nil)
	if err != nil {
		return nil, err
	}

	network := &types.NetworkListResp{}

	err = decodeBody(network, resp.Body)
	ensureCloseReader(resp)

	return network, err
}
