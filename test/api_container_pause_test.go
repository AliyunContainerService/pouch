package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIContainerPauseSuite is the test suite for container pause/unpause API.
type APIContainerPauseSuite struct{}

func init() {
	check.Suite(&APIContainerPauseSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerPauseSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TestPauseUnpauseOk tests a running container could be paused and unpaused.
func (suite *APIContainerPauseSuite) TestPauseUnpauseOk(c *check.C) {
	// must required
	cname := "TestPauseUnpauseOk"
	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// TODO: Add state check

	resp, err = request.Post("/containers/" + cname + "/unpause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	DelContainerForceOk(c, cname)
}

// TestNonExistingContainer tests pause a non-existing container return 404.
func (suite *APIContainerPauseSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestNotRunningContainer tests pausing a non-running container will return error.
func (suite *APIContainerPauseSuite) TestNotRunningContainer(c *check.C) {
	cname := "TestNotRunningContainer"
	CreateBusyboxContainerOk(c, cname)

	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	StartContainerOk(c, cname)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	resp, err = request.Post("/containers/" + cname + "/unpause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	StopContainerOk(c, cname)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	DelContainerForceOk(c, cname)
}
