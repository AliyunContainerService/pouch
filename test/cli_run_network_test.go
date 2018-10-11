package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunNetworkSuite is the test suite for run CLI.
type PouchRunNetworkSuite struct{}

func init() {
	check.Suite(&PouchRunNetworkSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunNetworkSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TestRunWithPing is to verify run container with network ping public website.
func (suite *PouchRunNetworkSuite) TestRunWithPing(c *check.C) {
	name := "TestRunWithPing"

	res := command.PouchRun("run", "--name", name,
		busyboxImage, "ping", "-c", "3", "www.taobao.com")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)
}
