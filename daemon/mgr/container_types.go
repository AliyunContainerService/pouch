package mgr

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/cri/stream/remotecommand"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/containerd/containerd/mount"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

var (
	// GCExecProcessTick is the time interval to trigger gc unused exec config,
	// time unit is minute.
	GCExecProcessTick = 5

	// MinMemory is minimal memory container should has.
	MinMemory int64 = 4194304

	// DefaultStatsInterval is the interval configured for stats.
	DefaultStatsInterval = time.Duration(time.Second)
)

var (
	// MemoryWarn is warning for flag --memory
	MemoryWarn = "Current Kernel does not support memory limit, discard --memory"

	// MemorySwapWarn is warning for flag --memory-swap
	MemorySwapWarn = "Current Kernel does not support memory swap, discard --memory-swap"

	// MemorySwappinessWarn is warning for flag --memory-swappiness
	MemorySwappinessWarn = "Current Kernel does not support memory swappiness , discard --memory-swappiness"

	//OOMKillWarn is warning for flag --oom-kill-disable
	OOMKillWarn = "Current Kernel does not support disable oom kill, discard --oom-kill-disable"

	// CpusetCpusWarn is warning for flag --cpuset-cpus
	CpusetCpusWarn = "Current Kernel does not support cpuset cpus, discard --cpuset-cpus"

	// CpusetMemsWarn is warning for flag --cpuset-mems
	CpusetMemsWarn = "Current Kernel does not support cpuset mems, discard --cpuset-mems"

	// CPUSharesWarn is warning for flag --cpu-shares
	CPUSharesWarn = "Current Kernel does not support cpu shares, discard --cpu-shares"

	// CPUQuotaWarn is warning for flag --cpu-quota
	CPUQuotaWarn = "Current Kernel does not support cpu quota, discard --cpu-quota"

	// CPUPeriodWarn is warning for flag --cpu-period
	CPUPeriodWarn = "Current Kernel does not support cpu period, discard --cpu-period"

	// BlkioWeightWarn is warning for flag --blkio-weight
	BlkioWeightWarn = "Current Kernel does not support blkio weight, discard --blkio-weight"

	// BlkioWeightDeviceWarn is warning for flag --blkio-weight-device
	BlkioWeightDeviceWarn = "Current Kernel does not support blkio weight device, discard --blkio-weight-device"

	// BlkioDeviceReadBpsWarn is warning for flag --device-read-bps
	BlkioDeviceReadBpsWarn = "Current Kernel does not support blkio device throttle read bps, discard --device-read-bps"

	// BlkioDeviceWriteBpsWarn is warning for flag --device-write-bps
	BlkioDeviceWriteBpsWarn = "Current Kernel does not support blkio device throttle write bps, discard --device-write-bps"

	// BlkioDeviceReadIOpsWarn is warning for flag --device-read-iops
	BlkioDeviceReadIOpsWarn = "Current Kernel does not support blkio device throttle read iops, discard --device-read-iops"

	// BlkioDeviceWriteIOpsWarn is warning for flag --device-write-iops
	BlkioDeviceWriteIOpsWarn = "Current Kernel does not support blkio device throttle, discard --device-write-iops"

	// PidsLimitWarn is warning for flag --pids-limit
	PidsLimitWarn = "Current Kernel does not support pids cgroup, discard --pids-limit"
)

const (
	// DefaultStopTimeout is the timeout (in seconds) for the syscall signal used to stop a container.
	DefaultStopTimeout = 10

	// RuntimeDir is specified name keeps runtime path script.
	RuntimeDir = "runtimes"
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
	All        bool
	Filter     map[string][]string
	FilterFunc ContainerFilter
}

// ContainerStatsConfig contains all configs on stats interface.
// This struct is only used in daemon side.
type ContainerStatsConfig struct {
	Stream    bool
	OutStream io.Writer
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

	// RootFSProvided is a flag to point the container is created by specify rootfs
	RootFSProvided bool

	// MountFS is used to mark the directory of mount overlayfs for pouch daemon to operate the image.
	MountFS string `json:"-"`
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
	imageConf, err := getconfig()
	if err != nil {
		return err
	}

	// If user specify the Entrypoint, no need to merge image's configuration.
	// Otherwise use the image's configuration to fill it.
	if len(c.Config.Entrypoint) == 0 {
		if len(c.Config.Cmd) == 0 {
			c.Config.Cmd = imageConf.Cmd
		}
		c.Config.Entrypoint = imageConf.Entrypoint
	}

	// ContainerConfig.Env is new, and the ImageConfig.Env is old
	newEnvSlice, err := mergeEnvSlice(c.Config.Env, imageConf.Env)
	if err != nil {
		return err
	}
	c.Config.Env = newEnvSlice
	if c.Config.WorkingDir == "" {
		c.Config.WorkingDir = imageConf.WorkingDir
	}

	// merge user from image image config.
	if c.Config.User == "" {
		c.Config.User = imageConf.User
	}

	// merge stop signal from image config.
	if c.Config.StopSignal == "" {
		c.Config.StopSignal = imageConf.StopSignal
	}

	// merge label from image image config, if same label key exist,
	// use container config.
	if imageConf.Labels != nil {
		if c.Config.Labels == nil {
			c.Config.Labels = make(map[string]string)
		}
		for k, v := range c.Config.Labels {
			imageConf.Labels[k] = v
		}
		c.Config.Labels = imageConf.Labels
	}

	// merge exposed ports from image config, if same label key exist,
	// use container config.
	if len(imageConf.ExposedPorts) > 0 {
		if c.Config.ExposedPorts == nil {
			c.Config.ExposedPorts = make(map[string]interface{})
		}
		for k, v := range imageConf.ExposedPorts {
			if _, exist := c.Config.ExposedPorts[k]; !exist {
				c.Config.ExposedPorts[k] = interface{}(v)
			}
		}
	}

	// merge volumes from image config.
	if len(imageConf.Volumes) > 0 {
		if c.Config.Volumes == nil {
			c.Config.Volumes = make(map[string]interface{})
		}
		for k, v := range imageConf.Volumes {
			if _, exist := c.Config.Volumes[k]; !exist {
				c.Config.Volumes[k] = interface{}(v)
			}
		}
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

// UnsetMergedDir unsets Snapshot MergedDir. Stop a container will
// delete the containerd container, the merged dir
// will also be deleted, so we should unset the
// container's MergedDir.
func (c *Container) UnsetMergedDir() {
	if c.Snapshotter == nil || c.Snapshotter.Data == nil {
		return
	}
	c.Snapshotter.Data["MergedDir"] = ""
}

// SetSnapshotterMeta sets snapshotter for container
func (c *Container) SetSnapshotterMeta(mounts []mount.Mount) {
	// TODO(ziren): now we only support overlayfs
	data := make(map[string]string, 0)
	for _, opt := range mounts[0].Options {
		if strings.HasPrefix(opt, "upperdir=") {
			data["UpperDir"] = strings.TrimPrefix(opt, "upperdir=")
		}
		if strings.HasPrefix(opt, "lowerdir=") {
			data["LowerDir"] = strings.TrimPrefix(opt, "lowerdir=")
		}
		if strings.HasPrefix(opt, "workdir=") {
			data["WorkDir"] = strings.TrimPrefix(opt, "workdir=")
		}
	}

	c.Snapshotter = &types.SnapshotterData{
		Name: "overlayfs",
		Data: data,
	}
}

// GetSpecificBasePath accepts a given path, look for whether the path is exist
// within container, if has, returns container base path like BaseFS, if not, return empty string
func (c *Container) GetSpecificBasePath(path string) string {
	logrus.Debugf("GetSpecificBasePath, snapshotter data: (%v)", c.Snapshotter.Data)

	// try lower and upper directory, since overlay filesystem support only.
	for _, key := range []string{"MergedDir", "UpperDir", "LowerDir"} {
		if prefixPath, ok := c.Snapshotter.Data[key]; ok && prefixPath != "" {
			for _, p := range strings.Split(prefixPath, ":") {
				absPath := filepath.Join(p, path)
				if utils.IsFileExist(absPath) {
					return absPath
				}
			}
		}
	}

	return ""
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
