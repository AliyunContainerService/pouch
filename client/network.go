package client

import (
	"github.com/alibaba/pouch/apis/types"
)

// NetworkList requests daemon to list all networks
func (client *APIClient) NetworkList() ([]types.Network, error) {
	resp, err := client.get("/networks/json", nil)
	if err != nil {
		return nil, err
	}

	networkList := []types.Network{}

	err = decodeBody(&networkList, resp.Body)
	ensureCloseReader(resp)

	return networkList, err
}
