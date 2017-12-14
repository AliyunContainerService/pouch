package ctrd

import (
	"context"

	"github.com/containerd/containerd/snapshots"
)

const defaultSnapshotterName = "overlayfs"

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
