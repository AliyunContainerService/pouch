package ctrd

import (
	"context"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/pkg/jsonstream"

	"github.com/containerd/containerd"
	containerdtypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
)

// APIClient defines common methods of containerd api client
type APIClient interface {
	ContainerAPIClient
	ImageAPIClient
	SnapshotAPIClient

	Version(ctx context.Context) (containerd.Version, error)
	Cleanup() error
}

// ContainerAPIClient provides access to containerd container features.
type ContainerAPIClient interface {
	// CreateContainer creates a containerd container and start process.
	CreateContainer(ctx context.Context, container *Container) error
	// DestroyContainer kill container and delete it.
	DestroyContainer(ctx context.Context, id string, timeout int64) (*Message, error)
	// ProbeContainer probe the container's status, if timeout <= 0, will block to receive message.
	ProbeContainer(ctx context.Context, id string, timeout time.Duration) *Message
	// ContainerPIDs returns the all processes's ids inside the container.
	ContainerPIDs(ctx context.Context, id string) ([]int, error)
	// ContainerPID returns the container's init process id.
	ContainerPID(ctx context.Context, id string) (int, error)
	// ContainerStats returns stats of the container.
	ContainerStats(ctx context.Context, id string) (*containerdtypes.Metric, error)
	// ExecContainer executes a process in container.
	ExecContainer(ctx context.Context, process *Process) error
	// RecoverContainer reload the container from metadata and watch it, if program be restarted.
	RecoverContainer(ctx context.Context, id string, io *containerio.IO) error
	// PauseContainer pause container.
	PauseContainer(ctx context.Context, id string) error
	// UnpauseContainer unpauses a container.
	UnpauseContainer(ctx context.Context, id string) error
	// ResizeContainer changes the size of the TTY of the init process running
	// in the container to the given height and width.
	ResizeContainer(ctx context.Context, id string, opts types.ResizeOptions) error
	// UpdateResources updates the configurations of a container.
	UpdateResources(ctx context.Context, id string, resources types.Resources) error
	// SetExitHooks specified the handlers of container exit.
	SetExitHooks(hooks ...func(string, *Message) error)
	// SetExecExitHooks specified the handlers of exec process exit.
	SetExecExitHooks(hooks ...func(string, *Message) error)
}

// ImageAPIClient provides access to containerd image features.
type ImageAPIClient interface {
	// GetImage returns containerd.Image by the given reference.
	GetImage(ctx context.Context, ref string) (containerd.Image, error)
	// ListImages returns the list of containerd.Image filtered by the given conditions.
	ListImages(ctx context.Context, filter ...string) ([]containerd.Image, error)
	// PullImage pulls image by the given reference.
	PullImage(ctx context.Context, ref string, authConfig *types.AuthConfig, stream *jsonstream.JSONStream) (containerd.Image, error)
	// RemoveImage removes the image by the given reference.
	RemoveImage(ctx context.Context, ref string) error
}

// SnapshotAPIClient provides access to containerd snapshot features
type SnapshotAPIClient interface {
	// CreateSnapshot creates a active snapshot with image's name and id.
	CreateSnapshot(ctx context.Context, id, ref string) error
	// GetSnapshot returns the snapshot's info by id.
	GetSnapshot(ctx context.Context, id string) (snapshots.Info, error)
	// RemoveSnapshot removes the snapshot by id.
	RemoveSnapshot(ctx context.Context, id string) error
	// GetMounts returns the mounts for the active snapshot transaction identified
	// by key.
	GetMounts(ctx context.Context, id string) ([]mount.Mount, error)
	// GetSnapshotUsage returns the resource usage of an active or committed snapshot
	// excluding the usage of parent snapshots.
	GetSnapshotUsage(ctx context.Context, id string) (snapshots.Usage, error)
	// WalkSnapshot walk all snapshots. For each snapshot, the function will be called.
	WalkSnapshot(ctx context.Context, fn func(context.Context, snapshots.Info) error) error
}
