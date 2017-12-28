package client

import "github.com/alibaba/pouch/apis/types"

// NetworkCreate creates a network.
func (client *APIClient) NetworkCreate(req *types.NetworkCreateConfig) (*types.NetworkCreateResp, error) {
	resp, err := client.post("/networks/create", nil, req)
	if err != nil {
		return nil, err
	}

	network := &types.NetworkCreateResp{}

	err = decodeBody(network, resp.Body)
	ensureCloseReader(resp)

	return network, err
}
