package client

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
)

// SystemPing shows whether server is ok.
func (client *APIClient) SystemPing(ctx context.Context) (string, error) {
	resp, err := client.get(ctx, "/_ping", nil, nil)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// SystemVersion requests daemon for system version.
func (client *APIClient) SystemVersion(ctx context.Context) (*types.SystemVersion, error) {
	resp, err := client.get(ctx, "/version", nil, nil)
	if err != nil {
		return nil, err
	}

	version := &types.SystemVersion{}
	err = decodeBody(version, resp.Body)
	ensureCloseReader(resp)

	return version, err
}

// SystemInfo requests daemon for system info.
func (client *APIClient) SystemInfo(ctx context.Context) (*types.SystemInfo, error) {
	resp, err := client.get(ctx, "/info", nil, nil)
	if err != nil {
		return nil, err
	}

	info := &types.SystemInfo{}
	err = decodeBody(info, resp.Body)
	ensureCloseReader(resp)

	return info, err
}

// RegistryLogin requests a registry server to login.
func (client *APIClient) RegistryLogin(ctx context.Context, auth *types.AuthConfig) (*types.AuthResponse, error) {
	resp, err := client.post(ctx, "/auth", nil, auth, nil)
	if err != nil || resp.StatusCode == http.StatusUnauthorized {
		return nil, err
	}

	authResp := &types.AuthResponse{}
	err = decodeBody(authResp, resp.Body)
	ensureCloseReader(resp)

	return authResp, err
}
