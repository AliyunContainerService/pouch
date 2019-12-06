package ctrd

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/pkg/log"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/mount"
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
func CurrentSnapshotterName(ctx context.Context) string {
	if v := GetSnapshotter(ctx); v != "" {
		return v
	}
	return currentSnapshotterName
}

// CreateSnapshot creates a active snapshot with image's name and id.
func (c *Client) CreateSnapshot(ctx context.Context, id, ref string) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	originalCtx := ctx
	ctx = leases.WithLease(ctx, wrapperCli.lease.ID)

	var (
		snName = CurrentSnapshotterName(ctx)
		snSrv  = wrapperCli.client.SnapshotService(snName)
	)

	image, err := wrapperCli.client.GetImage(ctx, ref)
	if err != nil {
		return err
	}

	diffIDs, err := image.RootFS(ctx)
	if err != nil {
		return err
	}
	parent := identity.ChainID(diffIDs).String()

	// NOTE: PouchContainer always unpacks image during pulling. But there
	// maybe crash or terminated by some reason. The image have been stored
	// in containerd without unpacking. And the following creating container
	// request will fail on preparing snapshot because there is no such
	// parent snapshotter. Based on this case, we should skip the not
	// found error and try to unpack it again.
	_, err = snSrv.Prepare(ctx, id, parent)
	if err == nil || !errdefs.IsNotFound(err) {
		return err
	}
	log.With(ctx).Warnf("checking unpack status for image %s on %s snapshotter...", image.Name(), snName)

	// check unpacked status
	unpacked, werr := image.IsUnpacked(ctx, snName)
	if werr != nil {
		log.With(ctx).Warnf("failed to check unpack status for image %s on %s snapshotter: %v", image.Name(), snName, werr)
		return werr
	}

	// if it is not unpacked, try to unpack it.
	if !unpacked {
		log.With(ctx).Warnf("the image %s doesn't unpack for %s snapshotter, try to unpack it...", image.Name(), snName)
		// NOTE: don't use pouchd lease id here because pouchd lease id
		// will hold the snapshotter forever, which means that the
		// snapshotter will not removed if we remove image.
		if werr = image.Unpack(originalCtx, snName); werr != nil {
			log.With(ctx).Warnf("failed to unpack for image %s on %s snapshotter: %v", image.Name(), snName, werr)
			return werr
		}

		// do it again.
		_, err = snSrv.Prepare(ctx, id, parent)
	}
	return err
}

// GetSnapshot returns the snapshot's info by id.
func (c *Client) GetSnapshot(ctx context.Context, id string) (snapshots.Info, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return snapshots.Info{}, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	service := wrapperCli.client.SnapshotService(CurrentSnapshotterName(ctx))
	defer service.Close()

	return service.Stat(ctx, id)
}

// RemoveSnapshot removes the snapshot by id.
func (c *Client) RemoveSnapshot(ctx context.Context, id string) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	service := wrapperCli.client.SnapshotService(CurrentSnapshotterName(ctx))
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

	service := wrapperCli.client.SnapshotService(CurrentSnapshotterName(ctx))
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

	service := wrapperCli.client.SnapshotService(CurrentSnapshotterName(ctx))
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
		snapshotter = CurrentSnapshotterName(ctx)
	}

	service := wrapperCli.client.SnapshotService(snapshotter)
	defer service.Close()

	return service.Walk(ctx, fn)
}
