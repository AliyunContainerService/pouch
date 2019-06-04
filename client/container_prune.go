package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils/filters"
)

// ContainerPrune requests the daemon to delete unused data
func (client *APIClient) ContainerPrune(ctx context.Context, filter map[string][]string) (*types.ContainerPruneResp, error) {

	var result types.ContainerPruneResp

	q := url.Values{}

	if len(filter) > 0 {
		fJSON, err := filters.ToURLParam(filter)
		if err != nil {
			return nil, err
		}
		q.Set("filters", fJSON)
	}

	resp, err := client.post(ctx, "/containers/prune", q, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &result, fmt.Errorf("error retrieving disk usage: %v", err)
	}

	return &result, nil
}
