package snapshot

import (
	"context"
	"sync"

	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	digest "github.com/opencontainers/go-digest"
)

type Mountable interface {
	// ID() string
	Mount() ([]mount.Mount, error)
	Release() error
}

type SnapshotterBase interface {
	Mounts(ctx context.Context, key string) (Mountable, error)
	Prepare(ctx context.Context, key, parent string, opts ...snapshots.Opt) error
	View(ctx context.Context, key, parent string, opts ...snapshots.Opt) (Mountable, error)

	Stat(ctx context.Context, key string) (snapshots.Info, error)
	Update(ctx context.Context, info snapshots.Info, fieldpaths ...string) (snapshots.Info, error)
	Usage(ctx context.Context, key string) (snapshots.Usage, error)
	Commit(ctx context.Context, name, key string, opts ...snapshots.Opt) error
	Remove(ctx context.Context, key string) error
	Walk(ctx context.Context, fn func(context.Context, snapshots.Info) error) error
	Close() error
}

// Snapshotter defines interface that any snapshot implementation should satisfy
type Snapshotter interface {
	Blobmapper
	SnapshotterBase
}

type Blobmapper interface {
	GetBlob(ctx context.Context, key string) (digest.Digest, digest.Digest, error)
	SetBlob(ctx context.Context, key string, diffID, blob digest.Digest) error
}

func FromContainerdSnapshotter(s snapshots.Snapshotter) SnapshotterBase {
	return &fromContainerd{Snapshotter: s}
}

type fromContainerd struct {
	snapshots.Snapshotter
}

func (s *fromContainerd) Mounts(ctx context.Context, key string) (Mountable, error) {
	mounts, err := s.Snapshotter.Mounts(ctx, key)
	if err != nil {
		return nil, err
	}
	return &staticMountable{mounts}, nil
}
func (s *fromContainerd) Prepare(ctx context.Context, key, parent string, opts ...snapshots.Opt) error {
	_, err := s.Snapshotter.Prepare(ctx, key, parent, opts...)
	return err
}
func (s *fromContainerd) View(ctx context.Context, key, parent string, opts ...snapshots.Opt) (Mountable, error) {
	mounts, err := s.Snapshotter.View(ctx, key, parent, opts...)
	if err != nil {
		return nil, err
	}
	return &staticMountable{mounts}, nil
}

type staticMountable struct {
	mounts []mount.Mount
}

func (m *staticMountable) Mount() ([]mount.Mount, error) {
	return m.mounts, nil
}

func (cm *staticMountable) Release() error {
	return nil
}

// NewContainerdSnapshotter converts snapshotter to containerd snapshotter
func NewContainerdSnapshotter(s Snapshotter) (snapshots.Snapshotter, func() error) {
	cs := &containerdSnapshotter{Snapshotter: s}
	return cs, cs.release
}

type containerdSnapshotter struct {
	mu        sync.Mutex
	releasers []func() error
	Snapshotter
}

func (cs *containerdSnapshotter) release() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	var err error
	for _, f := range cs.releasers {
		if err1 := f(); err != nil && err == nil {
			err = err1
		}
	}
	return err
}

func (cs *containerdSnapshotter) returnMounts(mf Mountable) ([]mount.Mount, error) {
	mounts, err := mf.Mount()
	if err != nil {
		return nil, err
	}
	cs.mu.Lock()
	cs.releasers = append(cs.releasers, mf.Release)
	cs.mu.Unlock()
	return mounts, nil
}

func (cs *containerdSnapshotter) Mounts(ctx context.Context, key string) ([]mount.Mount, error) {
	mf, err := cs.Snapshotter.Mounts(ctx, key)
	if err != nil {
		return nil, err
	}
	return cs.returnMounts(mf)
}

func (cs *containerdSnapshotter) Prepare(ctx context.Context, key, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	if err := cs.Snapshotter.Prepare(ctx, key, parent, opts...); err != nil {
		return nil, err
	}
	return cs.Mounts(ctx, key)
}
func (cs *containerdSnapshotter) View(ctx context.Context, key, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	mf, err := cs.Snapshotter.View(ctx, key, parent, opts...)
	if err != nil {
		return nil, err
	}
	return cs.returnMounts(mf)
}
