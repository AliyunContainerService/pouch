package containerplugin

import (
	"io"
)

// PreUpdate defines plugin point where receives a container update request, in this plugin point user
// could change the container update body passed-in by http request body.
func (c *contPlugin) PreUpdate(in io.ReadCloser) (io.ReadCloser, error) {
	// TODO: Implemented by the developer
	return in, nil
}

// PostUpdate called after update method successful,
// the method accepts the rootfs path and envs of container.
func (c *contPlugin) PostUpdate(rootfs string, env []string) error {
	// TODO: Implemented by the developer
	return nil
}
