package mgr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
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

func TestCopyOwnership(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "testCopyOwnerShip")
	if err != nil {
		t.Fatalf("failed to mk tmp dir: %v", err)
	}

	defer os.RemoveAll(tmpDir)

	sourcePath := filepath.Join(tmpDir, "sourceDir")

	err = os.Mkdir(sourcePath, 0611)
	if err != nil {
		t.Fatalf("failed to mkdir %s: %v", sourcePath, err)
	}

	err = os.Chown(sourcePath, 200, 300)
	if err != nil {
		t.Fatalf("failed to chown %s: %v", sourcePath, err)
	}

	dstPath := filepath.Join(tmpDir, "dstDir")
	err = os.Mkdir(dstPath, 0444)
	if err != nil {
		t.Fatalf("failed to mkdir %s: %v", dstPath, err)
	}

	err = copyOwnership(sourcePath, dstPath)
	if err != nil {
		t.Fatalf("copyOwnership from %s to %s got error: %v, expected nil", sourcePath, dstPath, err)
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", dstPath, err)
	}

	sysInfo, ok := dstInfo.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("failed to got %s stat info", dstPath)
	}

	if uint32(0611) != uint32(dstInfo.Mode().Perm()) {
		t.Fatalf("mode %d is not equal to %d", uint32(dstInfo.Mode().Perm()), uint32(0611))
	}

	if uint32(200) != sysInfo.Uid {
		t.Fatalf("Uid %d is not equal to %d", sysInfo.Uid, uint32(200))
	}

	if uint32(300) != sysInfo.Gid {
		t.Fatalf("Gid %d is not equal to %d", sysInfo.Gid, uint32(300))
	}
}
