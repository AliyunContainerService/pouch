package client

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
)

// ImageSearch requests daemon to search an image from registry.
func (client *APIClient) ImageSearch(ctx context.Context, term, registry, encodedAuth string) ([]types.SearchResultItem, error) {
	var results []types.SearchResultItem

	q := url.Values{}
	q.Set("term", term)
	q.Set("registry", registry)

	headers := map[string][]string{}
	if encodedAuth != "" {
		headers["X-Registry-Auth"] = []string{encodedAuth}
	}

	resp, err := client.post(ctx, "/images/search", q, nil, headers)

	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&results)
	return results, err
}
