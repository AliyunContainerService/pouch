package client

import (
	"context"
	"github.com/alibaba/pouch/apis/types"
)

// ImageList requests daemon to list all images
func (client *APIClient) ImageList(ctx context.Context) ([]types.ImageInfo, error) {
	resp, err := client.get(ctx, "/images/json", nil, nil)
	if err != nil {
		return nil, err
	}

	imageList := []types.ImageInfo{}

	err = decodeBody(&imageList, resp.Body)
	ensureCloseReader(resp)

	return imageList, err

}
