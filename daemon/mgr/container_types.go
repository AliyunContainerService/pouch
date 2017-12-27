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

// ContainerFilter defines a function to filter
// container in the store.
type ContainerFilter func(*ContainerMeta) bool

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

// ContainerMeta represents the container's meta data.
type ContainerMeta struct {

	// app armor profile
	AppArmorProfile string `json:"AppArmorProfile,omitempty"`

	// The arguments to the command being run
	Args []string `json:"Args"`

	// config
	Config *types.ContainerConfig `json:"Config,omitempty"`

	// The time the container was created
	Created string `json:"Created,omitempty"`

	// driver
	Driver string `json:"Driver,omitempty"`

	// exec ids
	ExecIds string `json:"ExecIDs,omitempty"`

	// graph driver
	GraphDriver *types.GraphDriverData `json:"GraphDriver,omitempty"`

	// host config
	HostConfig *types.HostConfig `json:"HostConfig,omitempty"`

	// hostname path
	HostnamePath string `json:"HostnamePath,omitempty"`

	// hosts path
	HostsPath string `json:"HostsPath,omitempty"`

	// The ID of the container
	ID string `json:"Id,omitempty"`

	// The container's image
	Image string `json:"Image,omitempty"`

	// log path
	LogPath string `json:"LogPath,omitempty"`

	// mount label
	MountLabel string `json:"MountLabel,omitempty"`

	// mounts
	Mounts []*types.MountPoint `json:"Mounts"`

	// name
	Name string `json:"Name,omitempty"`

	// network settings
	NetworkSettings *types.NetworkSettings `json:"NetworkSettings,omitempty"`

	Node interface{} `json:"Node,omitempty"`

	// The path to the command being run
	Path string `json:"Path,omitempty"`

	// process label
	ProcessLabel string `json:"ProcessLabel,omitempty"`

	// resolv conf path
	ResolvConfPath string `json:"ResolvConfPath,omitempty"`

	// restart count
	RestartCount int64 `json:"RestartCount,omitempty"`

	// The total size of all the files in this container.
	SizeRootFs int64 `json:"SizeRootFs,omitempty"`

	// The size of files that have been created or changed by this container.
	SizeRw int64 `json:"SizeRw,omitempty"`

	// state
	State *types.ContainerState `json:"State,omitempty"`
}

// Key returns container's id.
func (m *ContainerMeta) Key() string {
	return m.ID
}

// Container represents the container instance in runtime.
type Container struct {
	sync.Mutex
	meta       *ContainerMeta
	DetachKeys string
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
