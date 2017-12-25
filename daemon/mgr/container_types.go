package mgr

import (
	"net/http"
	"sync"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/meta"
)

const (
	// DefaultStopTimeout is the timeout (in seconds) for the syscall signal used to stop a container.
	DefaultStopTimeout = 10
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
	meta       *ContainerMeta
	DetachKeys string
}

// ContainerMeta wraps ContainerInfo and implements meta.Object interface.
// ContainerInfo is a struct only used in both client side and API server side.
// When request flow enters daemon's mgr side, all codes uses ContainerMeta to represent a container.
type ContainerMeta types.ContainerInfo

// Key returns container's id.
func (cm *ContainerMeta) Key() string {
	return cm.ID
}

// ToContainerInfo converts ContainerMeta to ContainerInfo
func (cm *ContainerMeta) ToContainerInfo() *types.ContainerInfo {
	containerInfo := types.ContainerInfo(*cm)
	return &containerInfo
}

// Key returns container's id.
func (c *Container) Key() string {
	return c.meta.ID
}

// ID returns container's id.
func (c *Container) ID() string {
	return c.meta.ID
}

// Image returns container's image name.
func (c *Container) Image() string {
	return c.meta.Config.Image
}

// Name returns container's name.
func (c *Container) Name() string {
	return c.meta.Name
}

// IsRunning returns container is running or not.
func (c *Container) IsRunning() bool {
	return c.meta.State.Status == types.StatusRunning
}

// IsStopped returns container is stopped or not.
func (c *Container) IsStopped() bool {
	return c.meta.State.Status == types.StatusStopped
}

// IsCreated returns container is created or not.
func (c *Container) IsCreated() bool {
	return c.meta.State.Status == types.StatusCreated
}

// IsPaused returns container is paused or not.
func (c *Container) IsPaused() bool {
	return c.meta.State.Status == types.StatusPaused
}

// Write writes container's meta data into meta store.
func (c *Container) Write(store *meta.Store) error {
	return store.Put(c.meta)
}

// StopTimeout returns the timeout (in seconds) used to stop the container.
func (c *Container) StopTimeout() int64 {
	if c.meta.Config.StopTimeout != nil {
		return *c.meta.Config.StopTimeout
	}
	return DefaultStopTimeout
}
