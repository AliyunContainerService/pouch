package containerplugin

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/hookplugins"
)

type contPlugin struct{}

func init() {
	hookplugins.RegisterContainerPlugin(&contPlugin{})
}

// PreCreate defines plugin point where receives a container create request, in this plugin point user
// could change the container create body passed-in by http request body
func (c *contPlugin) PreCreate(createConfig *types.ContainerCreateConfig) error {
	// TODO: Implemented by the developer
	return nil
}
