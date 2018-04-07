package client

import (
	"context"
	"net/url"
)

// ContainerRemove removes a container.
func (client *APIClient) ContainerRemove(ctx context.Context, name string, force bool) error {
	q := url.Values{}
	if force {
		q.Set("force", "true")
	}

	resp, err := client.delete(ctx, "/containers/"+name, q, nil)
	if err != nil {
		return err
	}
	ensureCloseReader(resp)
	return nil
}
