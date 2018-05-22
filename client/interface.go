package client

import (
	"bufio"
	"context"
	"io"
	"net"

	"github.com/alibaba/pouch/apis/types"
)

// CommonAPIClient defines common methods of api client
type CommonAPIClient interface {
	ContainerAPIClient
	ImageAPIClient
	VolumeAPIClient
	SystemAPIClient
	NetworkAPIClient
}

// ContainerAPIClient defines methods of Container client.
type ContainerAPIClient interface {
	ContainerCreate(ctx context.Context, config types.ContainerConfig, hostConfig *types.HostConfig, networkConfig *types.NetworkingConfig, containerName string) (*types.ContainerCreateResp, error)
	ContainerStart(ctx context.Context, name, detachKeys string) error
	ContainerStop(ctx context.Context, name, timeout string) error
	ContainerRemove(ctx context.Context, name string, options *types.ContainerRemoveOptions) error
	ContainerList(ctx context.Context, all bool) ([]*types.Container, error)
	ContainerAttach(ctx context.Context, name string, stdin bool) (net.Conn, *bufio.Reader, error)
	ContainerCreateExec(ctx context.Context, name string, config *types.ExecCreateConfig) (*types.ExecCreateResp, error)
	ContainerStartExec(ctx context.Context, execid string, config *types.ExecStartConfig) (net.Conn, *bufio.Reader, error)
	ContainerExecInspect(ctx context.Context, execid string) (*types.ContainerExecInspect, error)
	ContainerGet(ctx context.Context, name string) (*types.ContainerJSON, error)
	ContainerRename(ctx context.Context, id string, name string) error
	ContainerRestart(ctx context.Context, name string, timeout string) error
	ContainerPause(ctx context.Context, name string) error
	ContainerUnpause(ctx context.Context, name string) error
	ContainerUpdate(ctx context.Context, name string, config *types.UpdateConfig) error
	ContainerUpgrade(ctx context.Context, name string, config types.ContainerConfig, hostConfig *types.HostConfig) error
	ContainerTop(ctx context.Context, name string, arguments []string) (types.ContainerProcessList, error)
	ContainerLogs(ctx context.Context, name string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerResize(ctx context.Context, name, height, width string) error
}

// ImageAPIClient defines methods of Image client.
type ImageAPIClient interface {
	ImageList(ctx context.Context) ([]types.ImageInfo, error)
	ImageInspect(ctx context.Context, name string) (types.ImageInfo, error)
	ImagePull(ctx context.Context, name, tag, encodedAuth string) (io.ReadCloser, error)
	ImageRemove(ctx context.Context, name string, force bool) error
	ImageTag(ctx context.Context, image string, tag string) error
}

// VolumeAPIClient defines methods of Volume client.
type VolumeAPIClient interface {
	VolumeCreate(ctx context.Context, config *types.VolumeCreateConfig) (*types.VolumeInfo, error)
	VolumeRemove(ctx context.Context, name string) error
	VolumeInspect(ctx context.Context, name string) (*types.VolumeInfo, error)
	VolumeList(ctx context.Context) (*types.VolumeListResp, error)
}

// SystemAPIClient defines methods of System client.
type SystemAPIClient interface {
	SystemPing(ctx context.Context) (string, error)
	SystemVersion(ctx context.Context) (*types.SystemVersion, error)
	SystemInfo(ctx context.Context) (*types.SystemInfo, error)
	RegistryLogin(ctx context.Context, auth *types.AuthConfig) (*types.AuthResponse, error)
}

// NetworkAPIClient defines methods of Network client.
type NetworkAPIClient interface {
	NetworkCreate(ctx context.Context, req *types.NetworkCreateConfig) (*types.NetworkCreateResp, error)
	NetworkRemove(ctx context.Context, networkID string) error
	NetworkInspect(ctx context.Context, networkID string) (*types.NetworkInspectResp, error)
	NetworkList(ctx context.Context) ([]types.NetworkResource, error)
	NetworkConnect(ctx context.Context, network string, req *types.NetworkConnect) error
	NetworkDisconnect(ctx context.Context, networkID, containerID string, force bool) error
}
