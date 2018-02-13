package client

import (
	"context"
	"io"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
)

// ImageAPIClient defines methods of Image client.
type ImageAPIClient interface {
	ImageList(ctx context.Context) ([]types.ImageInfo, error)
	ImageInspect(ctx context.Context, name string) (types.ImageInfo, error)
	ImagePull(ctx context.Context, name, tag string) (io.ReadCloser, error)
	ImageRemove(ctx context.Context, name string, force bool) error
}

// ImageInspect requests daemon to inspect an image.
func (client *APIClient) ImageInspect(ctx context.Context, name string) (types.ImageInfo, error) {
	image := types.ImageInfo{}

	resp, err := client.get(ctx, "/images/"+name+"/json", nil)
	if err != nil {
		return image, err
	}

	defer ensureCloseReader(resp)
	err = decodeBody(&image, resp.Body)
	return image, err
}

// ImagePull requests daemon to pull an image from registry.
func (client *APIClient) ImagePull(ctx context.Context, name, tag string) (io.ReadCloser, error) {
	q := url.Values{}
	q.Set("fromImage", name)
	q.Set("tag", tag)

	resp, err := client.post(ctx, "/images/create", q, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ImageList requests daemon to list all images
func (client *APIClient) ImageList(ctx context.Context) ([]types.ImageInfo, error) {
	resp, err := client.get(ctx, "/images/json", nil)
	if err != nil {
		return nil, err
	}

	imageList := []types.ImageInfo{}

	err = decodeBody(&imageList, resp.Body)
	ensureCloseReader(resp)

	return imageList, err

}

// ImageRemove deletes an image.
func (client *APIClient) ImageRemove(ctx context.Context, name string, force bool) error {
	q := url.Values{}
	if force {
		q.Set("force", "true")
	}

	resp, err := client.delete(ctx, "/images/"+name, q)
	ensureCloseReader(resp)

	return err
}
