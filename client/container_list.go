package client

import (
	"context"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
)

// ContainerList returns the list of containers.
func (client *APIClient) ContainerList(ctx context.Context, all bool) ([]*types.Container, error) {
	q := url.Values{}
	if all {
		q.Set("all", "true")
	}

	resp, err := client.get(ctx, "/containers/json", q, nil)
	if err != nil {
		return nil, err
	}

	containers := []*types.Container{}
	err = decodeBody(&containers, resp.Body)
	ensureCloseReader(resp)

	return containers, err
}
