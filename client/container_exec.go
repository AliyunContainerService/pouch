package client

import (
	"bufio"
	"context"
	"net"
	"net/url"
	"strconv"

	"github.com/alibaba/pouch/apis/types"
)

// ContainerCreateExec creates exec process.
func (client *APIClient) ContainerCreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (*types.ExecCreateResp, error) {
	response, err := client.post(ctx, "/containers/"+name+"/exec", url.Values{}, config, nil)
	if err != nil {
		return nil, err
	}

	body := &types.ExecCreateResp{}
	err = decodeBody(body, response.Body)
	ensureCloseReader(response)

	return body, err
}

// ContainerStartExec starts exec process.
func (client *APIClient) ContainerStartExec(ctx context.Context, execID string, config *types.ExecStartConfig) (net.Conn, *bufio.Reader, error) {
	if config.Detach {
		_, err := client.post(ctx, "/exec/"+execID+"/start", url.Values{}, config, nil)
		return nil, nil, err
	}
	header := map[string][]string{
		"Content-Type": {"text/plain"},
	}

	return client.hijack(ctx, "/exec/"+execID+"/start", url.Values{}, config, header)
}

// ContainerExecInspect get exec info with a specified exec id.
func (client *APIClient) ContainerExecInspect(ctx context.Context, execID string) (*types.ContainerExecInspect, error) {
	resp, err := client.get(ctx, "/exec/"+execID+"/json", nil, nil)
	if err != nil {
		return nil, err
	}

	body := &types.ContainerExecInspect{}
	err = decodeBody(body, resp.Body)
	ensureCloseReader(resp)

	return body, err
}

// ContainerExecResize changes the size of the tty for an exec process running inside a container.
func (client *APIClient) ContainerExecResize(ctx context.Context, execID string, options types.ResizeOptions) error {
	query := url.Values{}
	query.Set("h", strconv.Itoa(int(options.Height)))
	query.Set("w", strconv.Itoa(int(options.Width)))

	resp, err := client.post(ctx, "/exec/"+execID+"/resize", query, nil, nil)
	ensureCloseReader(resp)
	return err
}
