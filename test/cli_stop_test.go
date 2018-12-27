package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchStopSuite is the test suite for stop CLI.
type PouchStopSuite struct{}

func init() {
	check.Suite(&PouchStopSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchStopSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchStopSuite) TearDownTest(c *check.C) {
}

// TestStopWorks tests "pouch stop" work.
func (suite *PouchStopSuite) TestStopWorks(c *check.C) {
	name := "stop-normal"

	command.PouchRun("create", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// test stop a created container
	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)

	// start the created container
	command.PouchRun("start", name).Assert(c, icmd.Success)

	// test stop a running container
	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
	// test stop a stopped container
	command.PouchRun("stop", name).Assert(c, icmd.Success)

	status, err := inspectFilter(name, ".State.Status")
	c.Assert(err, check.IsNil)
	c.Assert(status, check.Equals, "stopped")

	// test stop container with timeout(*seconds)
	command.PouchRun("start", name).Assert(c, icmd.Success)
	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)

	status, err = inspectFilter(name, ".State.Status")
	c.Assert(err, check.IsNil)
	c.Assert(status, check.Equals, "stopped")

	// test stop a paused container
	command.PouchRun("start", name).Assert(c, icmd.Success)
	command.PouchRun("pause", name).Assert(c, icmd.Success)
	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)

	status, err = inspectFilter(name, ".State.Status")
	c.Assert(err, check.IsNil)
	c.Assert(status, check.Equals, "stopped")
}

// TestStopInWrongWay tries to run create in wrong way.
func (suite *PouchStopSuite) TestStopInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "unknown container name", args: "unknown"},
		{name: "unknown flag", args: "-a"},
		{name: "Error: requires at least 1 arg(s), only received 0", args: ""},
	} {
		res := command.PouchRun("stop", "-t", "1", tc.args)
		c.Assert(res.Stderr(), check.NotNil, check.Commentf(tc.name))
	}
}

// TestStopMultiContainers tries to stop more than one container.
func (suite *PouchStopSuite) TestStopMultiContainers(c *check.C) {
	name1 := "TestStopMultiContainer-1"
	name2 := "TestStopMultiContainer-2"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name1, busyboxImage, "top").Assert(c, icmd.Success)
	command.PouchRun("run", "-d", "-m", "300M", "--name", name2, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name1)
	defer DelContainerForceMultyTime(c, name2)

	command.PouchRun("stop", "-t", "1", name1, name2).Assert(c, icmd.Success)

	// test if the container is already stopped
	status, err := inspectFilter(name1, ".State.Status")
	c.Assert(err, check.IsNil)
	c.Assert(status, check.Equals, "stopped")

	status, err = inspectFilter(name2, ".State.Status")
	c.Assert(err, check.IsNil)
	c.Assert(status, check.Equals, "stopped")
}

// TestStopPidValue ensure stopped container's pid is 0
func (suite *PouchStopSuite) TestStopPidValue(c *check.C) {
	name := "test-stop-pid-value"

	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// test stop a created container
	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)

	pid, err := inspectFilter(name, ".State.Pid")
	c.Assert(err, check.IsNil)
	c.Assert(pid, check.Equals, "0")
}

// TestAutoStopPidValue ensure stopped container's pid is 0
func (suite *PouchStopSuite) TestAutoStopPidValue(c *check.C) {
	name := "test-auto-stop-pid-value"

	command.PouchRun("run", "--name", name, busyboxImage, "echo", "hi").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	pid, err := inspectFilter(name, ".State.Pid")
	c.Assert(err, check.IsNil)
	c.Assert(pid, check.Equals, "0")
}
