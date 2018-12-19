package ctrd

import "context"

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
