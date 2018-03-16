package client

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

// ContainerCreate creates a new container based in the given configuration.
func (client *APIClient) ContainerCreate(ctx context.Context, config types.ContainerConfig, hostConfig *types.HostConfig, networkingConfig *types.NetworkingConfig, containerName string) (*types.ContainerCreateResp, error) {
	createConfig := types.ContainerCreateConfig{
		ContainerConfig:  config,
		HostConfig:       hostConfig,
		NetworkingConfig: networkingConfig,
	}

	q := url.Values{}
	if containerName != "" {
		q.Set("name", containerName)
	}

	resp, err := client.post(ctx, "/containers/create", q, createConfig, nil)
	if err != nil {
		return nil, err
	}

	container := &types.ContainerCreateResp{}

	err = decodeBody(container, resp.Body)
	ensureCloseReader(resp)

	return container, err
}

// ContainerStart starts a created container.
func (client *APIClient) ContainerStart(ctx context.Context, name, detachKeys string) error {
	q := url.Values{}
	if detachKeys != "" {
		q.Set("detachKeys", detachKeys)
	}

	resp, err := client.post(ctx, "/containers/"+name+"/start", q, nil, nil)
	ensureCloseReader(resp)

	return err
}

// ContainerStop stops a container.
func (client *APIClient) ContainerStop(ctx context.Context, name string, timeout string) error {
	q := url.Values{}
	q.Add("t", timeout)

	resp, err := client.post(ctx, "/containers/"+name+"/stop", q, nil, nil)
	ensureCloseReader(resp)

	return err
}

// ContainerRemove removes a container.
func (client *APIClient) ContainerRemove(ctx context.Context, name string, force bool) error {
	q := url.Values{}
	if force {
		q.Set("force", "true")
	}

	resp, err := client.delete(ctx, "/containers/"+name, q, nil)
	if err != nil {
		return err
	}
	ensureCloseReader(resp)
	return nil
}

// ContainerList returns the list of containers.
func (client *APIClient) ContainerList(ctx context.Context, all bool) ([]*types.Container, error) {
	q := url.Values{}
	if all {
		q.Set("all", "true")
	}

	resp, err := client.get(ctx, "/containers/json", q, nil)
	if err != nil {
		return nil, err
	}

	containers := []*types.Container{}
	err = decodeBody(&containers, resp.Body)
	ensureCloseReader(resp)

	return containers, err
}

// ContainerAttach attach a container
func (client *APIClient) ContainerAttach(ctx context.Context, name string, stdin bool) (net.Conn, *bufio.Reader, error) {
	q := url.Values{}
	if stdin {
		q.Set("stdin", "1")
	} else {
		q.Set("stdin", "0")
	}

	header := map[string][]string{
		"Content-Type": {"text/plain"},
	}

	return client.hijack(ctx, "/containers/"+name+"/attach", q, nil, header)
}

// ContainerCreateExec creates exec process.
func (client *APIClient) ContainerCreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (*types.ExecCreateResp, error) {
	response, err := client.post(ctx, "/containers/"+name+"/exec", url.Values{}, config, nil)
	if err != nil {
		return nil, err
	}

	body := &types.ExecCreateResp{}
	decodeBody(body, response.Body)
	ensureCloseReader(response)

	return body, nil
}

// ContainerStartExec starts exec process.
func (client *APIClient) ContainerStartExec(ctx context.Context, execid string, config *types.ExecStartConfig) (net.Conn, *bufio.Reader, error) {
	header := map[string][]string{
		"Content-Type": {"text/plain"},
	}

	return client.hijack(ctx, "/exec/"+execid+"/start", url.Values{}, config, header)
}

// ContainerGet returns the detailed information of container.
func (client *APIClient) ContainerGet(ctx context.Context, name string) (*types.ContainerJSON, error) {
	resp, err := client.get(ctx, "/containers/"+name+"/json", nil, nil)
	if err != nil {
		return nil, err
	}

	container := types.ContainerJSON{}
	err = decodeBody(&container, resp.Body)
	ensureCloseReader(resp)

	return &container, err
}

// ContainerRename renames a container.
func (client *APIClient) ContainerRename(ctx context.Context, id string, name string) error {
	q := url.Values{}
	q.Add("name", name)

	resp, err := client.post(ctx, "/containers/"+id+"/rename", q, nil, nil)
	ensureCloseReader(resp)

	return err
}

// ContainerRestart restarts a contianer.
func (client *APIClient) ContainerRestart(ctx context.Context, name string, time int) error {
	// TODO
	return nil
}

// ContainerPause pauses a container.
func (client *APIClient) ContainerPause(ctx context.Context, name string) error {
	resp, err := client.post(ctx, "/containers/"+name+"/pause", nil, nil, nil)
	ensureCloseReader(resp)

	return err
}

// ContainerUnpause unpauses a container.
func (client *APIClient) ContainerUnpause(ctx context.Context, name string) error {
	resp, err := client.post(ctx, "/containers/"+name+"/unpause", nil, nil, nil)
	ensureCloseReader(resp)

	return err
}

// ContainerUpdate updates the configurations of a container.
func (client *APIClient) ContainerUpdate(ctx context.Context, name string, config *types.UpdateConfig) error {
	resp, err := client.post(ctx, "/containers/"+name+"/update", url.Values{}, config, nil)
	ensureCloseReader(resp)

	return err

}

// ContainerUpgrade upgrade a container with new image and args.
func (client *APIClient) ContainerUpgrade(ctx context.Context, name string, config types.ContainerConfig, hostConfig *types.HostConfig) error {
	return nil
}

// ContainerTop shows process information from within a container.
func (client *APIClient) ContainerTop(ctx context.Context, name string, arguments []string) (types.ContainerProcessList, error) {
	response := types.ContainerProcessList{}
	query := url.Values{}
	if len(arguments) > 0 {
		query.Set("ps_args", strings.Join(arguments, " "))
	}

	resp, err := client.get(ctx, "/containers/"+name+"/top", query, nil)
	if err != nil {
		return response, err
	}

	err = decodeBody(&response, resp.Body)
	ensureCloseReader(resp)
	return response, err
}

// ContainerLogs return the logs generated by a container in an io.ReadCloser.
func (client *APIClient) ContainerLogs(ctx context.Context, name string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	query := url.Values{}
	if options.ShowStdout {
		query.Set("stdout", "1")
	}

	if options.ShowStderr {
		query.Set("stderr", "1")
	}

	if options.Since != "" {
		// TODO
	}

	if options.Until != "" {
		// TODO
	}

	if options.Timestamps {
		query.Set("timestamps", "1")
	}

	if options.Details {
		query.Set("details", "1")
	}

	if options.Follow {
		query.Set("follow", "1")
	}
	query.Set("tail", options.Tail)

	resp, err := client.get(ctx, "/containers/"+name+"/logs", query, nil)
	if err != nil {
		return nil, err
	}
	ensureCloseReader(resp)
	return resp.Body, nil
}
