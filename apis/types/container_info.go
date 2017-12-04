package types

// ContainerInfo accommodates container state and container configuration
type ContainerInfo struct {
	*ContainerState
	Config     *ContainerConfigWrapper
	ID         string
	Name       string
	DetachKeys string
}

// ContainerConfigWrapper is a type include config and host config of a container.
type ContainerConfigWrapper struct {
	*ContainerConfig
	HostConfig *HostConfig `json:"HostConfig,omitempty"`
}

// Key returns a key that stands for a container info which is the index in meta store.
func (c *ContainerInfo) Key() string {
	return c.ID
}
