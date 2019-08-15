package hookplugins

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
)

// VolumePlugin defines places where a plugin will be triggered in volume lifecycle
type VolumePlugin interface {
	// PreVolumeCreate defines plugin point where receives an volume create request, in this plugin point user
	// could change the volume create body passed-in by http request body
	PreVolumeCreate(context.Context, *types.VolumeCreateConfig) error
}

var volumePlugin VolumePlugin

// RegisterVolumePlugin is used to register the volume plugin.
func RegisterVolumePlugin(vp VolumePlugin) {
	volumePlugin = vp
}

// GetVolumePlugin returns the volume plugin.
func GetVolumePlugin() VolumePlugin {
	return volumePlugin
}
