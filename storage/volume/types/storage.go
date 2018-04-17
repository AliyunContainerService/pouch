package types

import (
	"time"

	"github.com/alibaba/pouch/storage/volume/types/meta"
)

// StorageSpec represents storage spec.
type StorageSpec struct {
	Type     string              `json:"type"`           // storage type
	ID       string              `json:"id,omitempty"`   // storage uid or unique name
	Name     string              `json:"name,omitempty"` // storage cluster name
	Key      string              `json:"key,omitempty"`  // storage access key
	API      string              `json:"api"`            // gateway address
	Address  string              `json:"address"`        // for ceph it's monitor ip:port,ip:port, pangu2's river master
	PoolSpec map[string]PoolSpec `json:"poolspec"`       // storage pool spec
}

// PoolSpec represents storage pool spec.
type PoolSpec struct {
	WriteBps   int64 `json:"iowbps"`         // storage write bytes per second
	ReadBps    int64 `json:"iorbps"`         // storage read bytes per second
	ReadIOps   int64 `json:"ioriops"`        // storage read io count per second
	WriteIOps  int64 `json:"iowiops"`        // storage write io count per second
	IOps       int64 `json:"ioiops"`         // storage total io count per second
	Capacity   int64 `json:"capacity"`       // storage capacity like: MB/GB/TB/PB
	Available  int64 `json:"available"`      // storage available like: MB/GB/TB/PB
	MaxVolumes int64 `json:"maxvolume"`      // max disks in this storage cluster
	NicSendBps int64 `json:"networkSendBps"` // nic card send bytes per second
	NicRecvBps int64 `json:"networkRecvBps"` // nic card recv bytes per second
}

// StorageStatus represents storage status.
type StorageStatus struct {
	Schedulable    bool       `json:"schedulable,omitempty"`
	LastUpdateTime *time.Time `json:"lastUpdateTime,omitempty"`
	HealthyStatus  string     `json:"message"`
}

// Storage represents storage struct.
type Storage struct {
	meta.ObjectMeta `json:",inline"`
	Spec            *StorageSpec   `json:"spec,omitempty"`
	Status          *StorageStatus `json:"status,omitempty"`
}

// Address return storage operate address.
func (s *Storage) Address() string {
	return s.Spec.Address
}

// SetAddress set storage operate address.
func (s *Storage) SetAddress(address string) {
	s.Spec.Address = address
}

// StorageList represents storage list type.
type StorageList struct {
	meta.ListMeta `json:",inline,omitempty"`
	Items         []Storage `json:"items,omitempty"`
}

// StorageID returns storage's uid.
func (s *Storage) StorageID() StorageID {
	return NewStorageID(s.GetUID())
}

// StorageID represents storage uid.
type StorageID struct {
	UID string
}

// NewStorageID is used to set uid.
func NewStorageID(uid string) StorageID {
	return StorageID{UID: uid}
}

// String returns storage uid.
func (si StorageID) String() string {
	return si.UID
}
