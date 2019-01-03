package ctrd

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

const (
	// TypeLabelKey is the key of label type=image in Snapshotter metadata Info.Labels, with which indicates the snapshot for image.
	TypeLabelKey = "type"

	// ImageType is the label value
	ImageType = "image"

	// SnapshotLabelContextKey is the key which is stored in context to transfer to containerd snapshotter
	SnapshotLabelContextKey = "containerd.io.snapshot.labels"
)

type snapshotterKey struct{}

// GetSnapshotter get snapshotter from context
func GetSnapshotter(ctx context.Context) string {
	snapshotter, _ := ctx.Value(snapshotterKey{}).(string)
	return snapshotter
}

// WithSnapshotter set snapshotter key for context
func WithSnapshotter(ctx context.Context, snapshotter string) context.Context {
	return context.WithValue(ctx, snapshotterKey{}, snapshotter)
}

// CleanSnapshotter cleans snapshotter key for context
func CleanSnapshotter(ctx context.Context) context.Context {
	if v, ok := ctx.Value(snapshotterKey{}).(string); ok && v != "" {
		return context.WithValue(ctx, snapshotterKey{}, "")
	}
	return ctx
}

// WithImageUnpack adds SnapshotLabelContextKey in context before creation of image snapshot.
// it should be called when image snapshot is on creation, such as pullImage, loadImage and commitImage.
func WithImageUnpack(ctx context.Context) context.Context {
	value := fmt.Sprintf("%s=%s", TypeLabelKey, ImageType)
	// store on the grpc headers so it gets picked up by any clients that
	// are using this.
	return withGRPCImageUnpackHeader(ctx, SnapshotLabelContextKey, value)
}

func withGRPCImageUnpackHeader(ctx context.Context, key string, value string) context.Context {
	// also store on the grpc headers so it gets picked up by any clients
	// that are using this.
	txheader := metadata.Pairs(key, value)
	md, ok := metadata.FromOutgoingContext(ctx) // merge with outgoing context.
	if !ok {
		md = txheader
	} else {
		// order ensures the latest is first in this list.
		md = metadata.Join(txheader, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
