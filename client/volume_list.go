package client

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
)

// VolumeList returns the list of volumes.
func (client *APIClient) VolumeList(ctx context.Context) (*types.VolumeListResp, error) {
	resp, err := client.get(ctx, "/volumes", nil, nil)
	if err != nil {
		return nil, err
	}

	volumeListResp := &types.VolumeListResp{}

	err = decodeBody(volumeListResp, resp.Body)
	ensureCloseReader(resp)

	return volumeListResp, err
}
