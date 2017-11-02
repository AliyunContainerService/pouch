package client

import (
	"net/url"
)

// ImagePull requests daemon to pull an image from registry.
func (cli *Client) ImagePull(name, tag string) error {
	q := url.Values{}
	q.Set("fromImage", name)
	q.Set("tag", tag)

	resp, err := cli.post("/images/create", q, nil)
	ensureCloseReader(resp)

	return err
}
