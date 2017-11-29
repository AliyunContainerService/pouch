package client

import (
	"bufio"
	"net"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
)

// ContainerCreate creates a new container based in the given configuration.
func (client *APIClient) ContainerCreate(config *types.ContainerConfig, hostConfig *types.HostConfig, containerName string) (*types.ContainerCreateResp, error) {
	createConfig := types.ContainerConfigWrapper{
		ContainerConfig: config,
		HostConfig:      hostConfig,
	}

	q := url.Values{}
	if containerName != "" {
		q.Set("name", containerName)
	}

	resp, err := client.post("/containers/create", q, createConfig)
	if err != nil {
		return nil, err
	}

	container := &types.ContainerCreateResp{}

	err = decodeBody(container, resp.Body)
	ensureCloseReader(resp)

	return container, err
}

// ContainerStart starts a created container.
func (client *APIClient) ContainerStart(name, detachKeys string) error {
	q := url.Values{}
	if detachKeys != "" {
		q.Set("detachKeys", detachKeys)
	}

	resp, err := client.post("/containers/"+name+"/start", q, nil)
	ensureCloseReader(resp)

	return err
}

// ContainerStop stops a container
func (client *APIClient) ContainerStop(name string) error {
	resp, err := client.post("/containers/"+name+"/stop", nil, nil)
	ensureCloseReader(resp)

	return err
}

// ContainerList returns the list of containers.
func (client *APIClient) ContainerList() ([]*types.Container, error) {
	resp, err := client.get("/containers/json", nil)
	if err != nil {
		return nil, err
	}

	containers := []*types.Container{}
	err = decodeBody(&containers, resp.Body)
	ensureCloseReader(resp)

	return containers, err
}

// ContainerAttach attach a container
func (client *APIClient) ContainerAttach(name string, stdin bool) (net.Conn, *bufio.Reader, error) {
	q := url.Values{}
	if stdin {
		q.Set("stdin", "1")
	} else {
		q.Set("stdin", "0")
	}

	header := map[string][]string{
		"Content-Type": {"text/plain"},
	}

	return client.hijack("/containers/"+name+"/attach", q, nil, header)
}

// ContainerCreateExec creates exec process.
func (client *APIClient) ContainerCreateExec(name string, config *types.ExecCreateConfig) (*types.ExecCreateResponse, error) {
	response, err := client.post("/containers/"+name+"/exec", url.Values{}, config)
	if err != nil {
		return nil, err
	}

	body := &types.ExecCreateResponse{}
	decodeBody(body, response.Body)
	ensureCloseReader(response)

	return body, nil
}

// ContainerStartExec starts exec process.
func (client *APIClient) ContainerStartExec(execid string, config *types.ExecStartConfig) (net.Conn, *bufio.Reader, error) {
	header := map[string][]string{
		"Content-Type": {"text/plain"},
	}

	return client.hijack("/exec/"+execid+"/start", url.Values{}, config, header)
}

func (client *APIClient) ContainerInfo(name string) (*types.ContainerInfo, error) {
	resp, err := client.get("/containers/"+name+"/json", nil)
	if err != nil {
		return nil, err
	}

	container := types.ContainerInfo{}
	err = decodeBody(&container, resp.Body)
	ensureCloseReader(resp)

	return &container, err
}
