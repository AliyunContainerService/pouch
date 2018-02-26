package client

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
)

// VolumeCreate creates a volume.
func (client *APIClient) VolumeCreate(ctx context.Context, config *types.VolumeCreateConfig) (*types.VolumeInfo, error) {
	resp, err := client.post(ctx, "/volumes/create", nil, config, nil)
	if err != nil {
		return nil, err
	}

	volume := &types.VolumeInfo{}

	err = decodeBody(volume, resp.Body)
	ensureCloseReader(resp)

	return volume, err
}

// VolumeRemove removes a volume.
func (client *APIClient) VolumeRemove(ctx context.Context, name string) error {
	resp, err := client.delete(ctx, "/volumes/"+name, nil, nil)
	ensureCloseReader(resp)

	return err
}

// VolumeInspect inspects a volume.
func (client *APIClient) VolumeInspect(ctx context.Context, name string) (*types.VolumeInfo, error) {
	resp, err := client.get(ctx, "/volumes/"+name, nil, nil)
	if err != nil {
		return nil, err
	}

	volume := &types.VolumeInfo{}

	err = decodeBody(volume, resp.Body)
	ensureCloseReader(resp)

	return volume, err
}

// VolumeList returns the list of volumes.
func (client *APIClient) VolumeList(ctx context.Context) (*types.VolumeListResp, error) {
	resp, err := client.get(ctx, "/volumes", nil, nil)

	volumeListResp := &types.VolumeListResp{}

	err = decodeBody(volumeListResp, resp.Body)
	ensureCloseReader(resp)

	return volumeListResp, err
}
