package hookplugins

import (
	"context"
	"io"

	"github.com/alibaba/pouch/apis/types"
	networktypes "github.com/alibaba/pouch/network/types"
)

// ContainerPlugin defines places where a plugin will be triggered in container lifecycle
type ContainerPlugin interface {
	// PreCreate defines plugin point where receives a container create request, in this plugin point user
	// could change the container create body passed-in by http request body
	PreCreate(context.Context, *types.ContainerCreateConfig) error

	// PreStart returns an array of priority and args which will pass to runc, the every priority
	// used to sort the pre start array that pass to runc, network plugin hook always has priority value 0.
	PreStart(context.Context, interface{}) ([]int, [][]string, error)

	// PreCreateEndpoint accepts the container id and env of this container, to update the config of container's endpoint.
	PreCreateEndpoint(context.Context, string, []string, *networktypes.Endpoint) error

	// PreUpdate defines plugin point where receives a container update request, in this plugin point user
	// could change the container update body passed-in by http request body
	PreUpdate(context.Context, io.ReadCloser) (io.ReadCloser, error)

	// PostUpdate called after update method successful,
	// the method accepts the rootfs path and envs of container
	PostUpdate(context.Context, string, []string) error
}

var containerPlugin ContainerPlugin

// RegisterContainerPlugin is used to register container plugin.
func RegisterContainerPlugin(cp ContainerPlugin) {
	containerPlugin = cp
}

// GetContainerPlugin returns the container plugin.
func GetContainerPlugin() ContainerPlugin {
	return containerPlugin
}
