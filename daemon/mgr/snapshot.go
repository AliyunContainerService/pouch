package mgr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pouch/ctrd"

	"github.com/containerd/containerd/snapshots"
	"github.com/sirupsen/logrus"
)

// Snapshot contains the information about the snapshot.
type snapshot struct {
	// Key is the key of the snapshot
	key string
	// Kind is the kind of the snapshot (active, commited, view)
	kind snapshots.Kind
	// Size is the size of the snapshot in bytes.
	size uint64
	// Inodes is the number of inodes used by the snapshot
	inodes uint64
	// Timestamp is latest update time (in nanoseconds) of the snapshot
	// information.
	timestamp int64
}

// snapshotStore stores all snapshots.
type snapshotStore struct {
	lock      sync.RWMutex
	snapshots map[string]snapshot
}

// newSnapshotStore create a new snapshot store.
func newSnapshotStore() *snapshotStore {
	return &snapshotStore{snapshots: make(map[string]snapshot)}
}

// add a snapshot into the store.
func (s *snapshotStore) add(sn snapshot) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.snapshots[sn.key] = sn
}

func (s *snapshotStore) get(key string) (snapshot, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if sn, ok := s.snapshots[key]; ok {
		return sn, nil
	}
	return snapshot{}, fmt.Errorf("failed to get %q in snapshot store", key)
}

// list lists all snapshots.
func (s *snapshotStore) list() []snapshot {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var snapshots []snapshot
	for _, sn := range s.snapshots {
		snapshots = append(snapshots, sn)
	}
	return snapshots
}

// delete deletes the snapshot with specified key.
func (s *snapshotStore) delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.snapshots, key)
}

// snapshotsSyncer syncs snapshot stats periodically.
type snapshotsSyncer struct {
	store      *snapshotStore
	client     ctrd.APIClient
	syncPeriod time.Duration
}

// newSnapshotsSyncer creates a snapshot syncer.
func newSnapshotsSyncer(store *snapshotStore, cli ctrd.APIClient, period time.Duration) *snapshotsSyncer {
	return &snapshotsSyncer{
		store:      store,
		client:     cli,
		syncPeriod: period,
	}
}

// start starts the snapshots syncer.
func (s *snapshotsSyncer) start() {
	tick := time.NewTicker(s.syncPeriod)
	go func() {
		defer tick.Stop()
		for {
			err := s.sync()
			if err != nil {
				logrus.Errorf("failed to sync snapshot stats: %v", err)
			}
			<-tick.C
		}
	}()
}

// sync updates the snapshots in the snapshot store.
func (s *snapshotsSyncer) sync() error {
	start := time.Now().UnixNano()
	var infos []snapshots.Info
	err := s.client.WalkSnapshot(context.Background(), func(ctx context.Context, info snapshots.Info) error {
		infos = append(infos, info)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk all snapshots: %v", err)
	}
	for _, info := range infos {
		sn, err := s.store.get(info.Name)
		if err == nil {
			// Only update timestamp for non-active snapshot.
			if sn.kind == info.Kind && sn.kind != snapshots.KindActive {
				sn.timestamp = time.Now().UnixNano()
				s.store.add(sn)
				continue
			}
		}
		// Get newest stats if the snapshot is new or active.
		sn = snapshot{
			key:       info.Name,
			kind:      info.Kind,
			timestamp: time.Now().UnixNano(),
		}
		usage, err := s.client.GetSnapshotUsage(context.Background(), info.Name)
		if err != nil {
			logrus.Errorf("failed to get usage for snapshot %q: %v", info.Name, err)
			continue
		}
		sn.size = uint64(usage.Size)
		sn.inodes = uint64(usage.Inodes)
		s.store.add(sn)
	}
	for _, sn := range s.store.list() {
		if sn.timestamp > start {
			continue
		}
		// Delete the snapshot stats if it's not updated this time.
		s.store.delete(sn.key)
	}

	return nil
}
