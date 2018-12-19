package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/alibaba/pouch/test/environment"

	"github.com/containerd/containerd/mount"
	"github.com/go-check/check"
)

// PouchSnapshotterSuite is the test suite for choose snapshotter feature.
type PouchSnapshotterSuite struct{}

func init() {
	check.Suite(&PouchSnapshotterSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchSnapshotterSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	environment.PruneAllContainers(apiClient)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchSnapshotterSuite) TearDownTest(c *check.C) {
}

// TestNotSetSnapshotter tests default snapshotter to run pouchd
func (suite *PouchSnapshotterSuite) TestNotSetSnapshotter(c *check.C) {
	dcfg, err := StartDefaultDaemon()
	c.Assert(err, check.IsNil)

	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
	c.Assert(result.ExitCode, check.Equals, 0)
	// clean busybox image
	defer RunWithSpecifiedDaemon(dcfg, "rmi", busyboxImage)

	// we can get snapshots dir from overlayfs dirs
	_, err = checkSnapshotsDir(dcfg.HomeDir, "overlayfs")
	c.Assert(err, check.IsNil)

}

// TestSetDefaultSnapshotter tests set default snapshotter driver to run pouchd
func (suite *PouchSnapshotterSuite) TestSetDefaultSnapshotter(c *check.C) {
	dcfg, err := StartDefaultDaemon("--snapshotter", "overlayfs")
	c.Assert(err, check.IsNil)

	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
	c.Assert(result.ExitCode, check.Equals, 0)
	// clean busybox image
	defer RunWithSpecifiedDaemon(dcfg, "rmi", busyboxImage)

	// we can get snapshots dir from overlayfs dirs
	_, err = checkSnapshotsDir(dcfg.HomeDir, "overlayfs")
	c.Assert(err, check.IsNil)
}

// TestOldSnapshotterNotClean tests old snapshotter driver not clean and then set a new one
func (suite *PouchSnapshotterSuite) TestOldSnapshotterNotClean(c *check.C) {
	dcfg, err := StartDefaultDaemon("--snapshotter", "overlayfs")
	c.Assert(err, check.IsNil)

	fileSystemInfo, err := mount.Lookup(dcfg.HomeDir)
	c.Assert(err, check.IsNil)

	if fileSystemInfo.FSType != "btrfs" {
		dcfg.KillDaemon()
		c.Skip("btrfs is not supported! Ignore test suite TestOldSnapshotterNotClean.")
	}

	result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
	c.Assert(result.ExitCode, check.Equals, 0)
	dcfg.KillDaemon()
	time.Sleep(10 * time.Second)

	_, err = StartDefaultDaemon("--snapshotter", "btrfs")
	fmt.Printf("start pouchd failed:%s", err.Error())
	c.Assert(err, check.NotNil)

	// clean image
	dcfg, err = StartDefaultDaemon("--snapshotter", "overlayfs")
	c.Assert(err, check.IsNil)

	result = RunWithSpecifiedDaemon(dcfg, "rmi", busyboxImage)
	c.Assert(result.ExitCode, check.Equals, 0)
	dcfg.KillDaemon()
}

// checkSnapshotsDir returns snapshots directory names by given snapshotter name
func checkSnapshotsDir(homeDir string, snapshotter string) ([]string, error) {
	const snapshotterPrefix = "io.containerd.snapshotter.v1."
	snapshotsDir := filepath.Join(homeDir, "containerd", "root", fmt.Sprintf("%s%s", snapshotterPrefix, snapshotter), "snapshots")

	_, err := os.Lstat(snapshotsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	children, err := ioutil.ReadDir(snapshotsDir)
	if err != nil {
		return nil, err
	}

	names := []string{}

	for _, f := range children {
		if !f.IsDir() {
			continue
		}

		names = append(names, f.Name())
	}

	return names, nil
}
