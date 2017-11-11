package client

import "github.com/alibaba/pouch/apis/types"

// VolumeCreate creates a volume
func (cli *Client) VolumeCreate(req *types.VolumeCreateRequest) (*types.VolumeInfo, error) {
	resp, err := cli.post("/volumes/create", nil, req)
	if err != nil {
		return nil, err
	}

	volume := &types.VolumeInfo{}

	err = decodeBody(volume, resp.Body)
	ensureCloseReader(resp)

	return volume, err
}

// VolumeRemove removes a volume
func (cli *Client) VolumeRemove(name string) error {
	resp, err := cli.delete("/volumes/"+name, nil)
	ensureCloseReader(resp)

	return err
}
