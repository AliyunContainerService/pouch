package client

import "github.com/alibaba/pouch/apis/types"

// VolumeCreate creates a volume
func (client *APIClient) VolumeCreate(req *types.VolumeCreateRequest) (*types.VolumeInfo, error) {
	resp, err := client.post("/volumes/create", nil, req)
	if err != nil {
		return nil, err
	}

	volume := &types.VolumeInfo{}

	err = decodeBody(volume, resp.Body)
	ensureCloseReader(resp)

	return volume, err
}

// VolumeRemove removes a volume
func (client *APIClient) VolumeRemove(name string) error {
	resp, err := client.delete("/volumes/"+name, nil)
	ensureCloseReader(resp)

	return err
}
