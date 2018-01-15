package client

import "github.com/alibaba/pouch/apis/types"

// VolumeCreate creates a volume
func (client *APIClient) VolumeCreate(config *types.VolumeCreateConfig) (*types.VolumeInfo, error) {
	resp, err := client.post("/volumes/create", nil, config)
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

// VolumeInfo returns a volume's information.
func (client *APIClient) VolumeInfo(name string) (*types.VolumeInfo, error) {
	resp, err := client.get("/volumes/"+name, nil)
	if err != nil {
		return nil, err
	}

	volume := &types.VolumeInfo{}

	err = decodeBody(volume, resp.Body)
	ensureCloseReader(resp)

	return volume, err
}
