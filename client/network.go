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

// NetworkRemove removes a network.
func (client *APIClient) NetworkRemove(networkID string) error {
	resp, err := client.delete("/networks/"+networkID, nil)
	if err != nil {
		return err
	}

	ensureCloseReader(resp)
	return err
}

//NetworkInspect inspects a network.
func (client *APIClient) NetworkInspect(networkID string) (*types.NetworkInspectResp, error) {
	resp, err := client.get("/networks/"+networkID, nil)
	if err != nil {
		return nil, err
	}

	network := &types.NetworkInspectResp{}

	err = decodeBody(network, resp.Body)
	ensureCloseReader(resp)

	return network, err
}
