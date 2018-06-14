package client

import (
	"context"
	"net/url"
)

// ContainerKill sends signal to a container.
func (client *APIClient) ContainerKill(ctx context.Context, name, signal string) error {
	q := url.Values{}
	q.Set("signal", signal)

	resp, err := client.post(ctx, "/containers/"+name+"/kill", q, nil, nil)
	if err != nil {
		return err
	}
	ensureCloseReader(resp)

	return nil
}
