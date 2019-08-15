package criplugin

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/hookplugins"
)

type criPlugin struct{}

func init() {
	hookplugins.RegisterCriPlugin(&criPlugin{})
}

// PreCreateContainer defines plugin point where receives a container create request, in this plugin point user
// could the container's config in cri interface.
func (c *criPlugin) PreCreateContainer(ctx context.Context, createConfig *types.ContainerCreateConfig, res interface{}) error {
	// TODO: Implemented by the developer
	return nil
}
