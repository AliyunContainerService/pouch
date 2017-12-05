package client

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
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
func (client *APIClient) ImageList() ([]types.Image, error) {
	resp, err := client.get("/images/json", nil)
	if err != nil {
		return nil, err
	}

	imageList := []types.Image{}

	err = decodeBody(&imageList, resp.Body)
	ensureCloseReader(resp)

	return imageList, err

}

// ImageRemove removes a image
func (client *APIClient) ImageRemove(name string, force bool) error {
	q := url.Values{}
	if force {
		q.Set("force", "true")
	}

	resp, err := client.delete("/images/"+name, q)
	if err != nil {
		return err
	}

	defer ensureCloseReader(resp)

	if resp.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}

	return nil
}
