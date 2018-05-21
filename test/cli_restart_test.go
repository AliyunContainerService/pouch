package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRestartSuite is the test suite for restart CLI.
type PouchRestartSuite struct{}

func init() {
	check.Suite(&PouchRestartSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRestartSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRestartSuite) TearDownTest(c *check.C) {
}

// TestPouchRestart is to verify the correctness of restarting a running container.
func (suite *PouchRestartSuite) TestPouchRestart(c *check.C) {
	name := "TestPouchRestart"

	res := command.PouchRun("run", "-d", "--cpu-share", "20", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("restart", "-t", "1", name)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output: %s, expected: %s", out, name)
	}
}

// TestPouchRestartStoppedContainer is to verify the correctness of restarting a stopped container.
// Pouch should be compatible with moby's API. Restarting a stopped container is allowed.
func (suite *PouchRestartSuite) TestPouchRestartStoppedContainer(c *check.C) {
	name := "TestPouchRestartStoppedContainer"

	res := command.PouchRun("create", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	command.PouchRun("restart", "-t", "1", name).Assert(c, icmd.Success)
}

// TestPouchRestartPausedContainer is to verify restart paused container
func (suite *PouchRestartSuite) TestPouchRestartPausedContainer(c *check.C) {
	name := "TestPouchRestartPausedContainer"

	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("pause", name).Assert(c, icmd.Success)

	command.PouchRun("restart", name).Assert(c, icmd.Success)
}

// TestPouchRestartMultiContainers is to verify the correctness of restarting more than one running container.
func (suite *PouchRestartSuite) TestPouchRestartMultiContainers(c *check.C) {
	containernames := []string{"TestPouchRestartMultiContainer-1", "TestPouchRestartMultiContainer-2"}
	for _, name := range containernames {
		res := command.PouchRun("run", "-d", "--cpu-share", "20", "--name", name, busyboxImage)
		defer DelContainerForceMultyTime(c, name)
		res.Assert(c, icmd.Success)
	}

	res := command.PouchRun("restart", "-t", "1", containernames[0], containernames[1])
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, containernames[0]) || !strings.Contains(out, containernames[1]) {
		c.Fatalf("unexpected output: %s, expected: %s\n%s", out, containernames[0], containernames[1])
	}
}
