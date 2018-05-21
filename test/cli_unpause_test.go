package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchUnpauseSuite is the test suite for unpause CLI.
type PouchUnpauseSuite struct{}

func init() {
	check.Suite(&PouchUnpauseSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchUnpauseSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchUnpauseSuite) TearDownTest(c *check.C) {
}

// TestUnpauseWorks tests "pouch unpause" work.
func (suite *PouchUnpauseSuite) TestUnpauseWorks(c *check.C) {
	containernames := []string{"bar1", "bar2"}
	for _, name := range containernames {
		command.PouchRun("create", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
		defer DelContainerForceMultyTime(c, name)

		command.PouchRun("start", name).Assert(c, icmd.Success)
	}

	command.PouchRun("pause", containernames[0]).Assert(c, icmd.Success)

	args := map[string]bool{
		// paused container
		containernames[0]: true,
		// running container
		containernames[1]: false,
		// not exist
		"bar3": false,
	}

	for arg, ok := range args {
		res := command.PouchRun("unpause", arg)
		if !ok {
			c.Assert(res.Stderr(), check.NotNil)
		} else {
			res.Assert(c, icmd.Success)
		}

	}
}

// TestUnpauseMultiContainers is to verify the correctness of unpausing more than one paused container.
func (suite *PouchUnpauseSuite) TestUnpauseMultiContainers(c *check.C) {
	containernames := []string{"bar1", "bar2"}
	for _, name := range containernames {
		command.PouchRun("create", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
		defer DelContainerForceMultyTime(c, name)

		command.PouchRun("start", name).Assert(c, icmd.Success)
	}

	res := command.PouchRun("pause", containernames[0], containernames[1])
	res.Assert(c, icmd.Success)

	res = command.PouchRun("unpause", containernames[0], containernames[1])
	res.Assert(c, icmd.Success)
}
