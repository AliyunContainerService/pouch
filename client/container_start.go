package client

import (
	"context"
	"net/url"
)

// ContainerStart starts a created container.
func (client *APIClient) ContainerStart(ctx context.Context, name, detachKeys string) error {
	q := url.Values{}
	if detachKeys != "" {
		q.Set("detachKeys", detachKeys)
	}

	resp, err := client.post(ctx, "/containers/"+name+"/start", q, nil, nil)
	ensureCloseReader(resp)

	return err
}
