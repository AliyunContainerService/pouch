package hookplugins

import (
	"github.com/alibaba/pouch/apis/types"
)

// CriPlugin defines places where a plugin will be triggered in CRI api lifecycle
type CriPlugin interface {
	// PreCreateContainer defines plugin point where receives a container create request, in this plugin point user
	// could update the container's config in cri interface.
	PreCreateContainer(*types.ContainerCreateConfig, interface{}) error
}

var criPlugin CriPlugin

// RegisterCriPlugin is used to register the cri plugin.
func RegisterCriPlugin(crip CriPlugin) {
	criPlugin = crip
}

// GetCriPlugin returns the cri plugin.
func GetCriPlugin() CriPlugin {
	return criPlugin
}
