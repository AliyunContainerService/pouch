package mgr

import (
	"testing"
	"time"

	"github.com/alibaba/pouch/pkg/errtypes"
	snapshot "github.com/containerd/containerd/snapshots"
	"github.com/stretchr/testify/assert"
)

func Test_SnapshotStore(t *testing.T) {
	snapshots := map[string]Snapshot{
		"key1": {
			Key:       "key1",
			Kind:      snapshot.KindActive,
			Size:      10,
			Inodes:    100,
			Timestamp: time.Now().UnixNano(),
		},
		"key2": {
			Key:       "key2",
			Kind:      snapshot.KindCommitted,
			Size:      20,
			Inodes:    200,
			Timestamp: time.Now().UnixNano(),
		},
		"key3": {
			Key:       "key3",
			Kind:      snapshot.KindView,
			Size:      0,
			Inodes:    0,
			Timestamp: time.Now().UnixNano(),
		},
	}

	s := NewSnapshotStore()

	t.Logf("should be able to add snapshot")
	for _, sn := range snapshots {
		s.Add(sn)
	}

	t.Logf("should be able to get snapshot")
	for id, sn := range snapshots {
		got, err := s.Get(id)
		assert.NoError(t, err)
		assert.Equal(t, sn, got)
	}

	t.Logf("should be able to list snapshot")
	sns := s.List()
	assert.Len(t, sns, 3)

	testKey := "key2"

	t.Logf("should be able to delete snapshot")
	s.Delete(testKey)
	sns = s.List()
	assert.Len(t, sns, 2)

	t.Logf("get should return empty struct and ErrNotExist after deletion")
	sn, err := s.Get(testKey)
	assert.Equal(t, Snapshot{}, sn)
	assert.Equal(t, errtypes.IsNotfound(err), true)
}
