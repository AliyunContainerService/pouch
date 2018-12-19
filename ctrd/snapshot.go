package ctrd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/snapshots"
	"github.com/opencontainers/image-spec/identity"
)

const (
	defaultSnapshotterName = "overlayfs"
)

var (
	currentSnapshotterName = defaultSnapshotterName
)

// SetSnapshotterName sets current snapshotter driver, it should be called only when daemon starts
func SetSnapshotterName(name string) {
	currentSnapshotterName = name
}

// CurrentSnapshotterName returns current snapshotter driver
func CurrentSnapshotterName() string {
	return currentSnapshotterName
}

// CreateSnapshot creates a active snapshot with image's name and id.
func (c *Client) CreateSnapshot(ctx context.Context, id, ref string, labels map[string]string) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}
	ctx = leases.WithLease(ctx, wrapperCli.lease.ID())

	image, err := wrapperCli.client.ImageService().Get(ctx, ref)
	if err != nil {
		return err
	}

	diffIDs, err := image.RootFS(ctx, wrapperCli.client.ContentStore(), platforms.Default())
	if err != nil {
		return err
	}

	opts := []snapshots.Opt{snapshots.WithLabels(labels)}

	parent := identity.ChainID(diffIDs).String()
	_, err = wrapperCli.client.SnapshotService(defaultSnapshotterName).Prepare(ctx, id, parent, opts...)
	return err
}

// GetSnapshot returns the snapshot's info by id.
func (c *Client) GetSnapshot(ctx context.Context, id string) (snapshots.Info, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return snapshots.Info{}, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	service := wrapperCli.client.SnapshotService(CurrentSnapshotterName())
	defer service.Close()

	return service.Stat(ctx, id)
}

// RemoveSnapshot removes the snapshot by id.
func (c *Client) RemoveSnapshot(ctx context.Context, id string) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	service := wrapperCli.client.SnapshotService(CurrentSnapshotterName())
	defer service.Close()

	return service.Remove(ctx, id)
}

// GetMounts returns the mounts for the active snapshot transaction identified
// by key.
func (c *Client) GetMounts(ctx context.Context, id string) ([]mount.Mount, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	service := wrapperCli.client.SnapshotService(CurrentSnapshotterName())
	defer service.Close()

	return service.Mounts(ctx, id)
}

// GetSnapshotUsage returns the resource usage of an active or committed snapshot
// excluding the usage of parent snapshots.
func (c *Client) GetSnapshotUsage(ctx context.Context, id string) (snapshots.Usage, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return snapshots.Usage{}, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	service := wrapperCli.client.SnapshotService(CurrentSnapshotterName())
	defer service.Close()

	return service.Usage(ctx, id)
}

// WalkSnapshot walk all snapshots in specific snapshotter. If not set specific snapshotter,
// it will be set to current snapshotter. For each snapshot, the function will be called.
func (c *Client) WalkSnapshot(ctx context.Context, snapshotter string, fn func(context.Context, snapshots.Info) error) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	// if not set specific snapshotter, set snapshotter to current snaphotter
	if snapshotter == "" {
		snapshotter = CurrentSnapshotterName()
	}

	service := wrapperCli.client.SnapshotService(snapshotter)
	defer service.Close()

	return service.Walk(ctx, fn)
}
