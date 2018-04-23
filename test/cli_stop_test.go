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

	command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)

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

	// FIXME: It's better if we use inspect to filter status.
	if out := res.Combined(); !strings.Contains(out, "Stopped") {
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

// TestStopMultiContainers tries to stop more than one container.
func (suite *PouchStopSuite) TestStopMultiContainers(c *check.C) {
	name1 := "TestStopMultiContainer-1"
	name2 := "TestStopMultiContainer-2"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name1, busyboxImage).Assert(c, icmd.Success)
	command.PouchRun("run", "-d", "-m", "300M", "--name", name2, busyboxImage).Assert(c, icmd.Success)

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
