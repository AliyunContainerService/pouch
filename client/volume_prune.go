package client

import (
	"context"
	"net/url"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
)

// VolumePrune delete unused volumes.
func (client *APIClient) VolumePrune(ctx context.Context, filter filters.Args) (*types.VolumePruneResp, error) {
	query := url.Values{}

	if filter.Len() > 0 {
		filtersJSON, err := filters.ToParam(filter)
		if err != nil {
			return nil, err
		}

		query.Set("filters", filtersJSON)
	}

	resp, err := client.post(ctx, "/volumes/prune", query, nil, nil)
	if err != nil {
		return nil, err
	}

	volumePruneResp := &types.VolumePruneResp{}

	err = decodeBody(volumePruneResp, resp.Body)
	ensureCloseReader(resp)

	return volumePruneResp, err
}
