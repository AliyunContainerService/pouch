package client

import (
	"io/ioutil"

	"github.com/alibaba/pouch/apis/types"
)

// SystemPing shows whether server is ok.
func (client *APIClient) SystemPing() (string, error) {
	resp, err := client.get("/_ping", nil)
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
func (client *APIClient) SystemVersion() (*types.SystemVersion, error) {
	resp, err := client.get("/version", nil)
	if err != nil {
		return nil, err
	}

	version := &types.SystemVersion{}
	err = decodeBody(version, resp.Body)
	ensureCloseReader(resp)

	return version, err
}

// SystemInfo requests daemon for system info.
func (client *APIClient) SystemInfo() (*types.SystemInfo, error) {
	resp, err := client.get("/info", nil)
	if err != nil {
		return nil, err
	}

	info := &types.SystemInfo{}
	err = decodeBody(info, resp.Body)
	ensureCloseReader(resp)

	return info, err
}
