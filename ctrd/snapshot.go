package ctrd

import (
	"context"

	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/snapshots"
	"github.com/opencontainers/image-spec/identity"
)

const defaultSnapshotterName = "overlayfs"

// CreateSnapshot creates a active snapshot with image's name and id.
func (c *Client) CreateSnapshot(ctx context.Context, id, ref string) error {
	ctx = leases.WithLease(ctx, c.lease.ID())

	image, err := c.client.ImageService().Get(ctx, ref)
	if err != nil {
		return err
	}

	diffIDs, err := image.RootFS(ctx, c.client.ContentStore(), platforms.Default())
	if err != nil {
		return err
	}

	parent := identity.ChainID(diffIDs).String()
	if _, err := c.client.SnapshotService(defaultSnapshotterName).Prepare(ctx, id, parent); err != nil {
		return err
	}
	return nil
}

// GetSnapshot returns the snapshot's info by id.
func (c *Client) GetSnapshot(ctx context.Context, id string) (snapshots.Info, error) {
	service := c.client.SnapshotService(defaultSnapshotterName)
	defer service.Close()

	return service.Stat(ctx, id)
}

// RemoveSnapshot removes the snapshot by id.
func (c *Client) RemoveSnapshot(ctx context.Context, id string) error {
	service := c.client.SnapshotService(defaultSnapshotterName)
	defer service.Close()

	return service.Remove(ctx, id)
}
