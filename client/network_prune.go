package client

import (
	"context"
	"net/url"
)

// NetworkPrune delete unused networks.
func (client *APIClient) NetworkPrune(ctx context.Context) ([]string, error) {
	resp, err := client.post(ctx, "/networks/prune", url.Values{}, nil, nil)
	if err != nil {
		return nil, err
	}

	result := []string{}
	err = decodeBody(&result, resp.Body)
	ensureCloseReader(resp)

	return result, err
}
