package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"
	"github.com/gotestyourself/gotestyourself/icmd"

	"github.com/go-check/check"
)

// PouchCheckpointSuite is the test suite for container create API.
type PouchCheckpointSuite struct{}

func init() {
	check.Suite(&PouchCheckpointSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchCheckpointSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	SkipIfFalse(c, environment.IsCRIUExist)
}

// TestCreate tests create checkpoint
func (suite *PouchCheckpointSuite) TestCheckpointCreate(c *check.C) {
	cname := "TestCheckpointCreate"

	command.PouchRun("run", "-d", "--name", cname, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	// create checkpoint cp0 leaving container running
	ret := command.PouchRun("checkpoint", "create", cname, "--leave-running", "cp0")
	ret.Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "cp0\n")

	// create a already exist checkpoint should fail
	ret = command.PouchRun("checkpoint", "create", cname, "cp0")
	c.Assert(util.PartialEqual(ret.Stderr(), "checkpoint cp0 is already exist"), check.IsNil)

	// create checkpoint cp1 leaving container exited
	ret = command.PouchRun("checkpoint", "create", cname, "cp1")
	ret.Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "cp1\n")
}

// TestCreate tests list checkpoint
func (suite *PouchCheckpointSuite) TestCheckpointList(c *check.C) {
	cname := "TestCheckpointList"

	command.PouchRun("run", "-d", "--name", cname, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	command.PouchRun("checkpoint", "create", cname, "--leave-running", "cp0").Assert(c, icmd.Success)
	ret := command.PouchRun("checkpoint", "ls", cname)
	ret.Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "cp0\n")

	command.PouchRun("checkpoint", "create", cname, "--leave-running", "cp1").Assert(c, icmd.Success)
	ret = command.PouchRun("checkpoint", "ls", cname)
	ret.Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "cp0\ncp1\n")
}

// TestCreate tests delete checkpoint
func (suite *PouchCheckpointSuite) TestCheckpointDelete(c *check.C) {
	cname := "TestCheckpointDelete"

	command.PouchRun("run", "-d", "--name", cname, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	command.PouchRun("checkpoint", "create", cname, "--leave-running", "cp0").Assert(c, icmd.Success)
	command.PouchRun("checkpoint", "create", cname, "cp1").Assert(c, icmd.Success)
	ret := command.PouchRun("checkpoint", "ls", cname)
	ret.Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "cp0\ncp1\n")

	// failed to delete a non-exist checkpoint
	ret = command.PouchRun("checkpoint", "rm", "no-exist")
	c.Assert(util.PartialEqual(ret.Stderr(), ""), check.IsNil)

	command.PouchRun("checkpoint", "rm", cname, "cp0").Assert(c, icmd.Success)
	command.PouchRun("checkpoint", "rm", cname, "cp1").Assert(c, icmd.Success)
	ret = command.PouchRun("checkpoint", "ls", cname)
	ret.Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "")
}
