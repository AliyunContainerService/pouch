package main

import (
	"encoding/json"
	"strings"

	"github.com/alibaba/pouch/apis/types"
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
	command.PouchRun("stop", name).Assert(c, icmd.Success)

	// start the created container
	command.PouchRun("start", name).Assert(c, icmd.Success)

	// test stop a running container
	command.PouchRun("stop", name).Assert(c, icmd.Success)
	// test stop a stopped container
	command.PouchRun("stop", name).Assert(c, icmd.Success)

	res := command.PouchRun("ps", "-a")

	// FIXME: It's better if we use inspect to filter status.
	if out := res.Combined(); !strings.Contains(out, "Stopped") {
		c.Fatalf("unexpected output %s expected Stopped\n", out)
	}

	// test stop container with timeout(*seconds)
	command.PouchRun("start", name).Assert(c, icmd.Success)
	command.PouchRun("stop", "-t", "3", name).Assert(c, icmd.Success)

	res = command.PouchRun("ps", "-a")

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(result[0].State.Status), check.Equals, "stopped")

	// test stop a paused container
	command.PouchRun("start", name).Assert(c, icmd.Success)
	command.PouchRun("pause", name).Assert(c, icmd.Success)
	command.PouchRun("stop", name).Assert(c, icmd.Success)

	output = command.PouchRun("inspect", name).Stdout()
	result = []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(result[0].State.Status), check.Equals, "stopped")
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
		res := command.PouchRun("stop", tc.args)
		c.Assert(res.Stderr(), check.NotNil, check.Commentf(tc.name))
	}
}

// TestStopMultiContainers tries to stop more than one container.
func (suite *PouchStopSuite) TestStopMultiContainers(c *check.C) {
	name1 := "TestStopMultiContainer-1"
	name2 := "TestStopMultiContainer-2"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name1, busyboxImage, "top").Assert(c, icmd.Success)
	command.PouchRun("run", "-d", "-m", "300M", "--name", name2, busyboxImage, "top").Assert(c, icmd.Success)

	command.PouchRun("stop", "-t", "3", name1, name2).Assert(c, icmd.Success)

	// test if the container is already stopped
	output := command.PouchRun("inspect", name1).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(result[0].State.Status), check.Equals, "stopped")

	output = command.PouchRun("inspect", name2).Stdout()
	result = []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(result[0].State.Status), check.Equals, "stopped")

}

// TestStopPidValue ensure stopped container's pid is 0
func (suite *PouchStopSuite) TestStopPidValue(c *check.C) {
	name := "test-stop-pid-value"

	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// test stop a created container
	command.PouchRun("stop", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].State.Pid, check.Equals, int64(0))
}

// TestAutoStopPidValue ensure stopped container's pid is 0
func (suite *PouchStopSuite) TestAutoStopPidValue(c *check.C) {
	name := "test-auto-stop-pid-value"

	command.PouchRun("run", "--name", name, busyboxImage, "echo", "hi").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].State.Pid, check.Equals, int64(0))
}
