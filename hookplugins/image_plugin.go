package hookplugins

import (
	"context"

	"github.com/containerd/containerd"
)

// ImagePlugin defines places where a plugin will be triggered in image operations
type ImagePlugin interface {
	PostPull(ctx context.Context, snapshotter string, image containerd.Image) error
}

var imagePlugin ImagePlugin

// RegisterImagePlugin is used to register container plugin.
func RegisterImagePlugin(p ImagePlugin) {
	imagePlugin = p
}

// GetImagePlugin returns the container plugin.
func GetImagePlugin() ImagePlugin {
	return imagePlugin
}
