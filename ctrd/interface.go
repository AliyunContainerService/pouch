package ctrd

import (
	"context"
	"io"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/pkg/jsonstream"

	"github.com/containerd/containerd"
	containerdtypes "github.com/containerd/containerd/api/types"
	ctrdmetaimages "github.com/containerd/containerd/images"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	digest "github.com/opencontainers/go-digest"
)

// APIClient defines common methods of containerd api client
type APIClient interface {
	ContainerAPIClient
	ImageAPIClient
	SnapshotAPIClient

	Version(ctx context.Context) (containerd.Version, error)
	Cleanup() error
	Plugins(ctx context.Context, filters []string) ([]Plugin, error)
	CheckSnapshotterValid(snapshotter string, allowMultiSnapshotter bool) error
}

// ContainerAPIClient provides access to containerd container features.
type ContainerAPIClient interface {
	// CreateContainer creates a containerd container and start process.
	CreateContainer(ctx context.Context, container *Container, checkpointDir string) error
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
	// ResizeContainer changes the size of the TTY of the exec process running
	// in the container to the given height and width.
	ResizeExec(ctx context.Context, id string, execid string, opts types.ResizeOptions) error
	// RecoverContainer reload the container from metadata and watch it, if program be restarted.
	RecoverContainer(ctx context.Context, id string, io *containerio.IO) error
	// PauseContainer pause container.
	PauseContainer(ctx context.Context, id string) error
	// UnpauseContainer unpauses a container.
	UnpauseContainer(ctx context.Context, id string) error
	// ResizeContainer changes the size of the TTY of the init process running
	// in the container to the given height and width.
	ResizeContainer(ctx context.Context, id string, opts types.ResizeOptions) error
	// WaitContainer waits until container's status is stopped.
	WaitContainer(ctx context.Context, id string) (types.ContainerWaitOKBody, error)
	// UpdateResources updates the configurations of a container.
	UpdateResources(ctx context.Context, id string, resources types.Resources) error
	// SetExitHooks specified the handlers of container exit.
	SetExitHooks(hooks ...func(string, *Message, func() error) error)
	// SetExecExitHooks specified the handlers of exec process exit.
	SetExecExitHooks(hooks ...func(string, *Message) error)
	// SetEventsHooks specified the methods to handle the containerd events.
	SetEventsHooks(hooks ...func(context.Context, string, string, map[string]string) error)
}

// ImageAPIClient provides access to containerd image features.
type ImageAPIClient interface {
	// CreateImageReference creates the image data into meta data in the containerd.
	CreateImageReference(ctx context.Context, img ctrdmetaimages.Image) (ctrdmetaimages.Image, error)
	// GetImage returns containerd.Image by the given reference.
	GetImage(ctx context.Context, ref string) (containerd.Image, error)
	// ListImages returns the list of containerd.Image filtered by the given conditions.
	ListImages(ctx context.Context, filter ...string) ([]containerd.Image, error)
	// PullImage fetches image content from the remote repository, and then unpacks into snapshotter
	PullImage(ctx context.Context, name string, refs []string, authConfig *types.AuthConfig, stream *jsonstream.JSONStream) (containerd.Image, error)
	// RemoveImage removes the image by the given reference.
	RemoveImage(ctx context.Context, ref string) error
	// ImportImage creates a set of images by tarstream.
	ImportImage(ctx context.Context, reader io.Reader, opts ...containerd.ImportOpt) ([]containerd.Image, error)
	// SaveImage saves image to tarstream
	SaveImage(ctx context.Context, exporter ctrdmetaimages.Exporter, ref string) (io.ReadCloser, error)
	// Commit commits an image from a container.
	Commit(ctx context.Context, config *CommitConfig) (digest.Digest, error)
	// PushImage pushes a image to registry
	PushImage(ctx context.Context, ref string, authConfig *types.AuthConfig, out io.Writer) error
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
	// WalkSnapshot walk all snapshots in specific snapshotter. If not set specific snapshotter,
	// it will be set to current snapshotter. For each snapshot, the function will be called.
	WalkSnapshot(ctx context.Context, snapshotter string, fn func(context.Context, snapshots.Info) error) error
	// CreateCheckpoint creates a checkpoint from a running container
	CreateCheckpoint(ctx context.Context, id string, checkpointDir string, exit bool) error
}
