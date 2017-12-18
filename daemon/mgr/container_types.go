package mgr

import (
	"net/http"
	"sync"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/meta"
)

type containerExecConfig struct {
	types.ExecCreateConfig

	// Save the container's id into exec config.
	ContainerID string
}

// AttachConfig wraps some infos of attaching.
type AttachConfig struct {
	Hijack  http.Hijacker
	Stdin   bool
	Stdout  bool
	Stderr  bool
	Upgrade bool
}

// ContainerRemoveOption wraps the container remove interface params.
type ContainerRemoveOption struct {
	Force  bool
	Volume bool
	Link   bool
}

// Container represents the container instance in runtime.
type Container struct {
	sync.Mutex
	meta *types.ContainerInfo
}

// ID returns container's id.
func (c *Container) ID() string {
	return c.meta.ID
}

// Name returns container's name.
func (c *Container) Name() string {
	return c.meta.Name
}

// IsRunning returns container is running or not.
func (c *Container) IsRunning() bool {
	return c.meta.Status == types.StatusRunning
}

// IsStopped returns container is stopped or not.
func (c *Container) IsStopped() bool {
	return c.meta.Status == types.StatusStopped
}

// IsCreated returns container is created or not.
func (c *Container) IsCreated() bool {
	return c.meta.Status == types.StatusCreated
}

// IsPaused returns container is paused or not.
func (c *Container) IsPaused() bool {
	return c.meta.Status == types.StatusPaused
}

// Write writes container's meta data into meta store.
func (c *Container) Write(store *meta.Store) error {
	return store.Put(c.meta)
}
