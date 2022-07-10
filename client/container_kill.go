package client

import (
	"context"
	"net/url"
)

// ContainerKill kill a container.
func (client *APIClient) ContainerKill(ctx context.Context, name string, signal string) error {
	q := url.Values{}
	q.Add("signal", signal)

	resp, err := client.post(ctx, "/containers/"+name+"/kill", q, nil, nil)
	ensureCloseReader(resp)

	return err
}
