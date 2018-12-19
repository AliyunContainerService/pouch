package ctrd

import (
	"context"
	"testing"
)

func TestSnapshotterKey(t *testing.T) {

	for _, tc := range []struct {
		key string
		ctx context.Context
	}{
		{
			key: "",
			ctx: context.TODO(),
		},
		{
			key: "foo",
			ctx: context.TODO(),
		},
		{
			key: "bar",
			ctx: context.TODO(),
		},
	} {
		tc.ctx = WithSnapshotter(tc.ctx, tc.key)
		if tc.key != GetSnapshotter(tc.ctx) {
			t.Fatalf("WithSnapshotter does not take effect, expect key %s, actual get %s", tc.key, GetSnapshotter(tc.ctx))
		}

		tc.ctx = CleanSnapshotter(tc.ctx)
		if "" != GetSnapshotter(tc.ctx) {
			t.Fatalf("CleanSnapshotter does not take effect, still get key %s", GetSnapshotter(tc.ctx))

		}
	}

}
