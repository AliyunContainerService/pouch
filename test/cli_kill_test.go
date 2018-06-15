package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchKillSuite is the test suite for kill CLI.
type PouchKillSuite struct{}

func init() {
	check.Suite(&PouchKillSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchKillSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchKillSuite) TeadDownTest(c *check.C) {
}

// TestKillWorks tests "pouch kill" work.
func (suite *PouchKillSuite) TestKillWorks(c *check.C) {
	name := "TestKillWorks"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("kill", name)
	res.Assert(c, icmd.Success)
	time.Sleep(250 * time.Millisecond)

	res = command.PouchRun("inspect", name)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(string(result[0].State.Status), check.Equals, "stopped")
}

// TestKillContainerWithSignal is to verify the correctness of sending signal to a container.
func (suite *PouchKillSuite) TestKillContainerWithSignal(c *check.C) {
	name := "TestKillContainerWithSignal"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("kill", "-s", "SIGWINCH", name)
	res.Assert(c, icmd.Success)
	time.Sleep(250 * time.Millisecond)

	res = command.PouchRun("inspect", name)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(string(result[0].State.Status), check.Equals, "running")
}

// TestKillContainerWithInvalidSignal is to verify the correctness of sending invalid signal to a container.
func (suite *PouchKillSuite) TestKillContainerWithInvalidSignal(c *check.C) {
	name := "TestKillContainerWithInvalidSignal"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("kill", "-s", "0", name)

	expectedError := "Invalid signal: 0"
	if out := res.Combined(); !strings.Contains(out, expectedError) {
		c.Fatalf("unexpected output %s expected %s", res.Stderr(), expectedError)
	}

	res = command.PouchRun("inspect", name)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(string(result[0].State.Status), check.Equals, "running")
}
