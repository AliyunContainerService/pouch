package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRmSuite is the test suite for pouch rm CLI.
type PouchRmSuite struct{}

func init() {
	check.Suite(&PouchRmSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRmSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TestContainerRmWithVolume tests remove container .
func (suite *PouchRmSuite) TestContainerRmWithVolume(c *check.C) {
	volumeName := "rmVolume-test-volume"
	containerName := "rmVolume-test"

	// create volume
	command.PouchRun("volume", "create", "-n", volumeName).Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
	}()

	ret := command.PouchRun("volume", "ls")
	expectVolumeNums := strings.Count(ret.Stdout(), "\n")

	// run container with volume
	command.PouchRun("run", "-d", "--name", containerName,
		"-v", volumeName+":/mnt",
		"-v", "/home",
		busyboxImage, "top").Assert(c, icmd.Success)

	command.PouchRun("rm", "-vf", containerName).Assert(c, icmd.Success)

	ret = command.PouchRun("volume", "ls")
	ret.Assert(c, icmd.Success)

	found := false
	volumeNums := 0
	for _, line := range strings.Split(ret.Stdout(), "\n") {
		if strings.Contains(line, volumeName) {
			found = true
		}
		volumeNums++
	}

	c.Assert(volumeNums, check.Equals, expectVolumeNums+1)
	c.Assert(found, check.Equals, true)
}
