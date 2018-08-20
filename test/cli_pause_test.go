package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchPauseSuite is the test suite for pause CLI.
type PouchPauseSuite struct{}

func init() {
	check.Suite(PouchPauseSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchPauseSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchPauseSuite) TearDownTest(c *check.C) {
}

// TestPauseWorks tests "pouch pause" work.
func (suite *PouchPauseSuite) TestPauseWorks(c *check.C) {
	name := "pause-normal"
	res := command.PouchRun("create", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	command.PouchRun("pause", name).Assert(c, icmd.Success)
}

// TestPauseInWrongWay tests run pause command in wrong way.
// TestPauseInWrongWay tests in four ways,
// pause with the container name missing
// pause a nonexistent container,
// pause with a nonexistent container name,
// and pause with an unwanted flag, respectively.
func (suite *PouchPauseSuite) TestPauseInWrongWay(c *check.C) {
	// generate a stopped container to test
	stoppedContainerName := "not running"
	res := command.PouchRun("create", "--name", stoppedContainerName, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, stoppedContainerName)
	res.Assert(c, icmd.Success)
	command.PouchRun("start", stoppedContainerName).Assert(c, icmd.Success)
	command.PouchRun("stop", stoppedContainerName).Assert(c, icmd.Success)

	for _, tc := range []struct {
		name          string
		args          string
		exepctedError string
	}{
		{
			name:          "missing container name",
			args:          "",
			exepctedError: "accepts at least 1 arg(s), received 0",
		},
		{
			name:          "nonexistent container name",
			args:          "non-existent",
			exepctedError: "not found",
		},
		{
			name:          "not running container name",
			args:          stoppedContainerName,
			exepctedError: "container's status is not running",
		},
		{
			name:          "unwanted flag",
			args:          "-m",
			exepctedError: "unknown shorthand flag",
		},
	} {
		res := command.PouchRun("pause", tc.args)
		c.Assert(res.Error, check.IsNil, check.Commentf(tc.name))

		if out := res.Combined(); !strings.Contains(out, tc.exepctedError) {
			c.Fatalf("unexpected output %s expected %s", res.Stderr(), tc.exepctedError)
		}
	}
}
