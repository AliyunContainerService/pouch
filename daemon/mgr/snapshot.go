package mgr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pouch/ctrd"

	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/containerd/containerd/snapshots"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Snapshot contains the information about the snapshot.
type Snapshot struct {
	// Key is the key of the snapshot
	Key string
	// Kind is the kind of the snapshot (active, committed, view)
	Kind snapshots.Kind
	// Size is the size of the snapshot in bytes.
	Size uint64
	// Inodes is the number of inodes used by the snapshot
	Inodes uint64
	// Timestamp is latest update time (in nanoseconds) of the snapshot
	// information.
	Timestamp int64
}

// SnapshotStore stores all snapshots.
type SnapshotStore struct {
	lock      sync.RWMutex
	snapshots map[string]Snapshot
}

// NewSnapshotStore create a new snapshot store.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{snapshots: make(map[string]Snapshot)}
}

// Add a snapshot into the store.
func (s *SnapshotStore) Add(sn Snapshot) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.snapshots[sn.Key] = sn
}

// Get returns the snapshot with specified key. Returns error if the
// snapshot doesn't exist.
func (s *SnapshotStore) Get(key string) (Snapshot, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if sn, ok := s.snapshots[key]; ok {
		return sn, nil
	}
	return Snapshot{}, errors.Wrapf(errtypes.ErrNotfound, "snapshot %s", key)
}

// List lists all snapshots.
func (s *SnapshotStore) List() []Snapshot {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var snapshots []Snapshot
	for _, sn := range s.snapshots {
		snapshots = append(snapshots, sn)
	}
	return snapshots
}

// Delete deletes the snapshot with specified key.
func (s *SnapshotStore) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.snapshots, key)
}

// SnapshotsSyncer syncs snapshot stats periodically.
type SnapshotsSyncer struct {
	store      *SnapshotStore
	client     ctrd.APIClient
	syncPeriod time.Duration
}

// newSnapshotsSyncer creates a snapshot syncer.
func newSnapshotsSyncer(store *SnapshotStore, cli ctrd.APIClient, period time.Duration) *SnapshotsSyncer {
	return &SnapshotsSyncer{
		store:      store,
		client:     cli,
		syncPeriod: period,
	}
}

// Start starts the snapshots syncer.
func (s *SnapshotsSyncer) Start() {
	tick := time.NewTicker(s.syncPeriod)
	go func() {
		defer tick.Stop()
		for {
			err := s.Sync()
			if err != nil {
				// TODO track the error and report it to the monitor or something.
				logrus.Errorf("failed to sync snapshot stats: %v", err)
			}
			<-tick.C
		}
	}()
}

// Sync updates the snapshots in the snapshot store.
func (s *SnapshotsSyncer) Sync() error {
	start := time.Now().UnixNano()
	var infos []snapshots.Info
	err := s.client.WalkSnapshot(context.Background(), "", func(ctx context.Context, info snapshots.Info) error {
		infos = append(infos, info)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk all snapshots: %v", err)
	}
	for _, info := range infos {
		sn, err := s.store.Get(info.Name)
		if err == nil {
			// Only update timestamp for non-active snapshot.
			if sn.Kind == info.Kind && sn.Kind != snapshots.KindActive {
				sn.Timestamp = time.Now().UnixNano()
				s.store.Add(sn)
				continue
			}
		}
		// Get newest stats if the snapshot is new or active.
		sn = Snapshot{
			Key:       info.Name,
			Kind:      info.Kind,
			Timestamp: time.Now().UnixNano(),
		}
		usage, err := s.client.GetSnapshotUsage(context.Background(), info.Name)
		if err != nil {
			logrus.Errorf("failed to get usage for snapshot %q: %v", info.Name, err)
			continue
		}
		sn.Size = uint64(usage.Size)
		sn.Inodes = uint64(usage.Inodes)
		s.store.Add(sn)
	}
	for _, sn := range s.store.List() {
		if sn.Timestamp > start {
			continue
		}
		// Delete the snapshot stats if it's not updated this time.
		// When remove a container,you also need to remove snapshot.
		// However, SnapshotStore will not be notified.
		// So wo need to delete snapshots from SnapshotStore that doesn't exist actually.
		s.store.Delete(sn.Key)
	}

	return nil
}
