package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchTopSuite is the test suite for top CLI.
type PouchTopSuite struct{}

func init() {
	check.Suite(&PouchTopSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchTopSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchTopSuite) TearDownTest(c *check.C) {
}

// TestTopStoppedContainer is to verify the correctness of top a stopped container.
func (suite *PouchTopSuite) TestTopStoppedContainer(c *check.C) {
	name := "TestTopStoppedContainer"

	command.PouchRun("create", "-m", "300M", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("top", name)
	c.Assert(res.Error, check.NotNil)

	expectString := "container is not running, can not execute top command"
	if out := res.Combined(); !strings.Contains(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestTopContainer is to verify the correctness of pouch top command.
func (suite *PouchTopSuite) TestTopContainer(c *check.C) {
	name := "TestTopContainer"

	command.PouchRun("run", "-m", "300M", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("top", name)
	c.Assert(res.Error, check.IsNil)

	expectString := "UID     PID      PPID     C    STIME    TTY    TIME        CMD"
	if out := res.Combined(); !strings.Contains(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}
