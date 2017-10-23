package ctrd

import (
	"context"

	"github.com/containerd/containerd"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// NewDefaultSpec new a template spec with default.
func NewDefaultSpec(ctx context.Context) (*specs.Spec, error) {
	return containerd.GenerateSpec(ctx, nil, nil)
}
