package driver

import (
	"errors"

	"github.com/alibaba/pouch/plugins"
)

// the following const variables is the protocol of remote volume driver.
const (
	// remoteVolumeCreateService is the service path of creating volume.
	remoteVolumeCreateService = "/VolumeDriver.Create"

	// remoteVolumeRemoveService is the service path of removing volume.
	remoteVolumeRemoveService = "/VolumeDriver.Remove"

	// remoteVolumeMountService is the service path of mounting volume.
	remoteVolumeMountService = "/VolumeDriver.Mount"

	// remoteVolumeUnmountService is the service path of unmounting volume.
	remoteVolumeUnmountService = "/VolumeDriver.Unmount"

	// remoteVolumePathService is the service path of getting volume's mountpath.
	remoteVolumePathService = "/VolumeDriver.Path"

	// remoteVolumeGetService is the service path of getting volume.
	remoteVolumeGetService = "/VolumeDriver.Get"

	// remoteVolumeListService is the service path of listing all volumes.
	remoteVolumeListService = "/VolumeDriver.List"

	// remoteVolumeListService is the service path of getting the volume driver' capabilities.
	remoteVolumeCapabilitiesService = "/VolumeDriver.Capabilities"
)

// remoteDriverProxy is a remote driver proxy.
type remoteDriverProxy struct {
	Name   string
	client *plugins.PluginClient
}

// Create creates a volume.
func (proxy *remoteDriverProxy) Create(name string, opts map[string]string) error {
	var req = remoteVolumeCreateReq{
		Name: name,
		Opts: opts,
	}

	var resp remoteVolumeCreateResp

	if err := proxy.client.CallService(remoteVolumeCreateService, &req, &resp, true); err != nil {
		return err
	}

	if resp.Err != "" {
		return errors.New(resp.Err)
	}

	return nil
}

// Remove deletes a volume.
func (proxy *remoteDriverProxy) Remove(name string) error {
	var req = remoteVolumeRemoveReq{
		Name: name,
	}

	var resp remoteVolumeCreateResp

	if err := proxy.client.CallService(remoteVolumeRemoveService, &req, &resp, true); err != nil {
		return err
	}

	if resp.Err != "" {
		return errors.New(resp.Err)
	}

	return nil
}

// Mount mounts a volume.
func (proxy *remoteDriverProxy) Mount(name, id string) (string, error) {
	var req = remoteVolumeMountReq{
		Name: name,
		ID:   id,
	}

	var resp remoteVolumeMountResp

	if err := proxy.client.CallService(remoteVolumeMountService, &req, &resp, true); err != nil {
		return "", err
	}

	if resp.Err != "" {
		return "", errors.New(resp.Err)
	}

	return resp.Mountpoint, nil
}

/// Umount unmounts a volume.
func (proxy *remoteDriverProxy) Unmount(name, id string) error {
	var req = remoteVolumeUnmountReq{
		Name: name,
		ID:   id,
	}

	var resp remoteVolumeUnmountResp

	if err := proxy.client.CallService(remoteVolumeUnmountService, &req, &resp, true); err != nil {
		return err
	}

	if resp.Err != "" {
		return errors.New(resp.Err)
	}

	return nil
}

// Path returns the mount path.
func (proxy *remoteDriverProxy) Path(name string) (string, error) {
	var req = remoteVolumePathReq{
		Name: name,
	}

	var resp remoteVolumePathResp

	if err := proxy.client.CallService(remoteVolumePathService, &req, &resp, true); err != nil {
		return "", err
	}

	if resp.Err != "" {
		return "", errors.New(resp.Err)
	}

	return resp.Mountpoint, nil
}

// Get returns the remote volume.
func (proxy *remoteDriverProxy) Get(name string) (*remoteVolumeDesc, error) {
	var req = remoteVolumeGetReq{
		Name: name,
	}

	var resp remoteVolumeGetResp

	if err := proxy.client.CallService(remoteVolumeGetService, &req, &resp, true); err != nil {
		return nil, err
	}

	if resp.Err != "" {
		return nil, errors.New(resp.Err)
	}

	return resp.Volume, nil
}

// List returns all remote volumes.
func (proxy *remoteDriverProxy) List() ([]*remoteVolumeDesc, error) {
	var req remoteVolumeListReq
	var resp remoteVolumeListResp

	if err := proxy.client.CallService(remoteVolumeListService, &req, &resp, true); err != nil {
		return nil, err
	}

	if resp.Err != "" {
		return nil, errors.New(resp.Err)
	}

	return resp.Volumes, nil
}

// Capabilities returns the driver capabilities.
func (proxy *remoteDriverProxy) Capabilities() (*remoteVolumeCapability, error) {
	var req remoteVolumeCapabilitiesReq
	var resp remoteVolumeCapabilitiesResp

	if err := proxy.client.CallService(remoteVolumeCapabilitiesService, &req, &resp, true); err != nil {
		return nil, err
	}

	if resp.Err != "" {
		return nil, errors.New(resp.Err)
	}

	return resp.Capabilities, nil
}
