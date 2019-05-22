package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
)

// ImagePrune requests daemon to delete unused images
func (client *APIClient) ImagePrune(ctx context.Context, filter filters.Args) (types.ImagePruneResp, error) {
	var result types.ImagePruneResp

	query := url.Values{}

	if filter.Len() > 0 {
		filtersJSON, err := filters.ToParam(filter)
		if err != nil {
			return result, err
		}

		query.Set("filters", filtersJSON)
	}

	resp, err := client.post(ctx, "/images/prune", query, nil, nil)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("error retrieving disk usage: %v", err)
	}

	return result, err
}
