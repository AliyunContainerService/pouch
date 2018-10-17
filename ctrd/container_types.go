package ctrd

import (
	"github.com/alibaba/pouch/daemon/containerio"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Container wraps container's info. there have two kind of containers now:
// One is created by pouch: first using image to create snapshot,
// then create container by specifying the snapshot;
// The other is create container by specify container rootfs, we use `RootFSProvided` flag to mark it,
type Container struct {
	ID         string
	Image      string
	Runtime    string
	Labels     map[string]string
	IO         *containerio.IO
	Spec       *specs.Spec
	SnapshotID string

	// BaseFS is rootfs used by containerd container
	BaseFS string

	// RootFSProvided is a flag to point the container is created by specifying rootfs
	RootFSProvided bool
}

// Process wraps exec process's info.
type Process struct {
	ContainerID string
	ExecID      string
	IO          *containerio.IO
	P           *specs.Process
}
