package client

import (
	"io"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
)

// ImageAPIClient defines methods of Image client.
type ImageAPIClient interface {
	ImageList() ([]types.ImageInfo, error)
	ImageInspect(name string) (types.ImageInfo, error)
	ImagePull(name, tag string) (io.ReadCloser, error)
	ImageRemove(name string, force bool) error
}

// ImageInspect requests daemon to inspect an image.
func (client *APIClient) ImageInspect(name string) (types.ImageInfo, error) {
	image := types.ImageInfo{}

	resp, err := client.get("/images/"+name+"/json", nil)
	if err != nil {
		return image, err
	}

	defer ensureCloseReader(resp)
	err = decodeBody(&image, resp.Body)
	return image, err
}

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

// ImageRemove deletes an image.
func (client *APIClient) ImageRemove(name string, force bool) error {
	q := url.Values{}
	if force {
		q.Set("force", "true")
	}

	resp, err := client.delete("/images/"+name, q)
	ensureCloseReader(resp)

	return err
}
