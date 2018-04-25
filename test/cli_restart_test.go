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

	command.PouchRun("run", "-d", "--cpu-share", "20", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("restart", "-t", "1", name)
	c.Assert(res.Error, check.IsNil)

	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output: %s, expected: %s", out, name)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestPouchRestartStoppedContainer is to verify the correctness of restarting a stopped container.
// Pouch should be compatible with moby's API. Restarting a stopped container is allowed.
func (suite *PouchRestartSuite) TestPouchRestartStoppedContainer(c *check.C) {
	name := "TestPouchRestartStoppedContainer"

	command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("restart", "-t", "1", name)
	c.Assert(res.Error, check.IsNil)

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}
