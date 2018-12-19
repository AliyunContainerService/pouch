package imageplugin

import (
	"context"

	"github.com/alibaba/pouch/hookplugins"

	"github.com/containerd/containerd"
)

type imagePlugin struct{}

func init() {
	hookplugins.RegisterImagePlugin(&imagePlugin{})
}

// PostPull is called after pull image
func (i *imagePlugin) PostPull(ctx context.Context, snapshotter string, image containerd.Image) error {
	// TODO: Implemented by the developer
	return nil
}
