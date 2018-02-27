package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchStopSuite is the test suite for help CLI.
type PouchStopSuite struct{}

func init() {
	check.Suite(&PouchStopSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchStopSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchStopSuite) TearDownTest(c *check.C) {
}

// TestStopWorks tests "pouch stop" work.
func (suite *PouchStopSuite) TestStopWorks(c *check.C) {
	name := "stop-normal"

	command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)
	defer command.PouchRun("rm", "-f", name)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	command.PouchRun("stop", name).Assert(c, icmd.Success)

	res := command.PouchRun("ps", "-a")

	// FIXME: It's better if we use inspect to filter status.
	if out := res.Combined(); !strings.Contains(out, "stopped") {
		c.Fatalf("unexpected output %s expected Stopped\n", out)
	}

	// test stop container with timeout(*seconds)
	command.PouchRun("start", name).Assert(c, icmd.Success)
	command.PouchRun("stop", "-t", "3", name).Assert(c, icmd.Success)

	res = command.PouchRun("ps", "-a")

	// FIXME: It's better if we use inspect to filter status.
	if out := res.Combined(); !strings.Contains(out, "stopped") {
		c.Fatalf("unexpected output %s expected Stopped\n", out)
	}
}

// TestStopInWrongWay tries to run create in wrong way.
func (suite *PouchStopSuite) TestStopInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "unknown container name", args: "unknown"},
		{name: "unknown flag", args: "-a"},

		// TODO: should add the following cases if ready
		// {name: "missing container name", args: ""},
	} {
		res := command.PouchRun("stop", tc.args)
		c.Assert(res.Error, check.NotNil, check.Commentf(tc.name))
	}
}
