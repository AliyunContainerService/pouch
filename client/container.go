package client

import (
	"net/url"

	"github.com/alibaba/pouch/apis/types"
)

// ContainerCreate creates a new container based in the given configuration.
func (cli *Client) ContainerCreate(config *types.ContainerConfig, hostConfig *types.HostConfig, containerName string) (*types.ContainerCreateResp, error) {
	createConfig := types.ContainerConfigWrapper{
		ContainerConfig: config,
		HostConfig:      hostConfig,
	}

	q := url.Values{}
	if containerName != "" {
		q.Set("name", containerName)
	}

	resp, err := cli.post("/containers/create", q, createConfig)
	if err != nil {
		return nil, err
	}

	container := &types.ContainerCreateResp{}

	err = decodeBody(container, resp.Body)
	ensureCloseReader(resp)

	return container, err
}

// ContainerStart starts a created container.
func (cli *Client) ContainerStart(name, detachKeys string) error {
	q := url.Values{}
	if detachKeys != "" {
		q.Set("detachKeys", detachKeys)
	}

	resp, err := cli.post("/containers/"+name+"/start", q, nil)
	ensureCloseReader(resp)

	return err
}
