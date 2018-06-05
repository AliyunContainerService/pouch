package mgr

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/cri/stream/remotecommand"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	// GCExecProcessTick is the time interval to trigger gc unused exec config,
	// time unit is minute.
	GCExecProcessTick = 5
)

const (
	// DefaultStopTimeout is the timeout (in seconds) for the syscall signal used to stop a container.
	DefaultStopTimeout = 10
)

// ContainerFilter defines a function to filter
// container in the store.
type ContainerFilter func(*Container) bool

// ContainerExecConfig is the config a process exec.
type ContainerExecConfig struct {
	// ExecID identifies the ID of this exec
	ExecID string

	// contains the config of this exec
	types.ExecCreateConfig

	// Save the container's id into exec config.
	ContainerID string

	// ExitCode records the exit code of a exec process.
	ExitCode int64

	// Running represents whether the exec process is running inside container.
	Running bool

	// Error represents the exec process response error.
	Error error

	// WaitForClean means exec process can be removed.
	WaitForClean bool
}

// AttachConfig wraps some infos of attaching.
type AttachConfig struct {
	Stdin  bool
	Stdout bool
	Stderr bool

	// For IO backend like http, we need to mux stdout & stderr
	// if terminal is disabled.
	// But for other IO backend, it is not necessary.
	// So we should make it configurable.
	MuxDisabled bool

	// Attach using http.
	Hijack  http.Hijacker
	Upgrade bool

	// Attach using pipe.
	Pipe *io.PipeWriter

	// Attach using streams.
	Streams *remotecommand.Streams

	// Attach to the container to get its log.
	CriLogFile *os.File
}

// ContainerListOption wraps the container list interface params.
type ContainerListOption struct {
	All bool
}

// Container represents the container's meta data.
type Container struct {
	sync.Mutex

	// app armor profile
	AppArmorProfile string `json:"AppArmorProfile,omitempty"`

	// seccomp profile
	SeccompProfile string `json:"SeccompProfile,omitempty"`

	// no new privileges
	NoNewPrivileges bool `json:"NoNewPrivileges,omitempty"`

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

	// Snapshotter, GraphDriver is same, keep both
	// just for compatibility
	// snapshotter informations of container
	Snapshotter *types.SnapshotterData `json:"Snapshotter,omitempty"`

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

	// BaseFS
	BaseFS string `json:"BaseFS, omitempty"`

	// Escape keys for detach
	DetachKeys string
}

// Key returns container's id.
func (c *Container) Key() string {
	c.Lock()
	defer c.Unlock()
	return c.ID
}

// Write writes container's meta data into meta store.
func (c *Container) Write(store *meta.Store) error {
	return store.Put(c)
}

// StopTimeout returns the timeout (in seconds) used to stop the container.
func (c *Container) StopTimeout() int64 {
	c.Lock()
	defer c.Unlock()
	if c.Config.StopTimeout != nil {
		return *c.Config.StopTimeout
	}
	return DefaultStopTimeout
}

func (c *Container) merge(getconfig func() (v1.ImageConfig, error)) error {
	c.Lock()
	defer c.Unlock()
	config, err := getconfig()
	if err != nil {
		return err
	}

	// If user specify the Entrypoint, no need to merge image's configuration.
	// Otherwise use the image's configuration to fill it.
	if len(c.Config.Entrypoint) == 0 {
		if len(c.Config.Cmd) == 0 {
			c.Config.Cmd = config.Cmd
		}
		c.Config.Entrypoint = config.Entrypoint
	}
	if c.Config.Env == nil {
		c.Config.Env = config.Env
	} else {
		c.Config.Env = append(c.Config.Env, config.Env...)
	}
	if c.Config.WorkingDir == "" {
		c.Config.WorkingDir = config.WorkingDir
	}

	return nil
}

// FormatStatus format container status
func (c *Container) FormatStatus() (string, error) {
	c.Lock()
	defer c.Unlock()
	var status string

	switch c.State.Status {
	case types.StatusRunning, types.StatusPaused:
		start, err := time.Parse(utils.TimeLayout, c.State.StartedAt)
		if err != nil {
			return "", err
		}

		startAt, err := utils.FormatTimeInterval(start.UnixNano())
		if err != nil {
			return "", err
		}

		status = "Up " + startAt
		if c.State.Status == types.StatusPaused {
			status += "(paused)"
		}

	case types.StatusStopped, types.StatusExited:
		finish, err := time.Parse(utils.TimeLayout, c.State.FinishedAt)
		if err != nil {
			return "", err
		}

		finishAt, err := utils.FormatTimeInterval(finish.UnixNano())
		if err != nil {
			return "", err
		}

		//FIXME: if stop status is needed ?
		exitCode := c.State.ExitCode
		if c.State.Status == types.StatusStopped {
			status = fmt.Sprintf("Stopped (%d) %s", exitCode, finishAt)
		}
		if c.State.Status == types.StatusExited {
			status = fmt.Sprintf("Exited (%d) %s", exitCode, finishAt)
		}
	}

	if status == "" {
		return string(c.State.Status), nil
	}

	return status, nil
}

// ContainerRestartPolicy represents the policy is used to manage container.
type ContainerRestartPolicy types.RestartPolicy

// IsNone returns the container don't need to be restarted or not.
func (p ContainerRestartPolicy) IsNone() bool {
	return p.Name == "" || p.Name == "no"
}

// IsAlways returns the container need to be restarted or not.
func (p ContainerRestartPolicy) IsAlways() bool {
	return p.Name == "always"
}
