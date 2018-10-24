package volumeplugin

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/hookplugins"
)

type volumePlugin struct{}

func init() {
	hookplugins.RegisterVolumePlugin(&volumePlugin{})
}

// PreVolumeCreate defines plugin point where receives an volume create request, in this plugin point user
// could change the volume create body passed-in by http request body
func (v *volumePlugin) PreVolumeCreate(config *types.VolumeCreateConfig) error {
	// TODO: Implemented by the developer
	return nil
}
