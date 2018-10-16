package plugins

import (
	"github.com/alibaba/pouch/apis/types"
)

// VolumePlugin defines places where a plugin will be triggered in volume lifecycle
type VolumePlugin interface {
	// PreCreate defines plugin point where receives an volume create request, in this plugin point user
	// could change the volume create body passed-in by http request body
	PreVolumeCreate(*types.VolumeCreateConfig) error
}
