package client

import (
	"io"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
)

// ImagePull requests daemon to pull an image from registry.
func (client *APIClient) ImagePull(name, tag string) (io.ReadCloser, error) {
	q := url.Values{}
	q.Set("fromImage", name)
	q.Set("tag", tag)

	resp, err := client.post("/images/create", q, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ImageList requests daemon to list all images
func (client *APIClient) ImageList() ([]types.ImageInfo, error) {
	resp, err := client.get("/images/json", nil)
	if err != nil {
		return nil, err
	}

	imageList := []types.ImageInfo{}

	err = decodeBody(&imageList, resp.Body)
	ensureCloseReader(resp)

	return imageList, err

}
