package mgr

import (
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func TestSortMountPoint(t *testing.T) {
	mounts := []*types.MountPoint{
		{
			Destination: "/home/admin/log",
			Source:      "/pouch/volume1",
		},
		{
			Destination: "/var/lib",
			Source:      "/pouch/volume2",
		},
		{
			Destination: "/home/admin",
			Source:      "/pouch/volume3",
		},
		{
			Destination: "/home",
			Source:      "/pouch/volume4",
		},
		{
			Destination: "/opt",
			Source:      "/pouch/volume5",
		},
		{
			Destination: "/var/log",
			Source:      "/pouch/volume6",
		},
	}

	expectMounts := []*types.MountPoint{
		{
			Destination: "/opt",
			Source:      "/pouch/volume5",
		},
		{
			Destination: "/home",
			Source:      "/pouch/volume4",
		},
		{
			Destination: "/var/lib",
			Source:      "/pouch/volume2",
		},
		{
			Destination: "/var/log",
			Source:      "/pouch/volume6",
		},
		{
			Destination: "/home/admin",
			Source:      "/pouch/volume3",
		},
		{
			Destination: "/home/admin/log",
			Source:      "/pouch/volume1",
		},
	}

	mounts = sortMountPoint(mounts)
	for i := range mounts {
		if mounts[i].Destination != expectMounts[i].Destination ||
			mounts[i].Source != expectMounts[i].Source {
			t.Fatalf("got unexpect sort: %v, expect: %v", *mounts[i], *expectMounts[i])
		}
	}
}
