package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerPauseSuite is the test suite for container pause/unpause API.
type APIContainerPauseSuite struct{}

func init() {
	check.Suite(&APIContainerPauseSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerPauseSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestPauseUnpauseOk tests a running container could be paused and unpaused.
func (suite *APIContainerPauseSuite) TestPauseUnpauseOk(c *check.C) {
	// must required
	cname := "TestPauseUnpauseOk"
	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	// pause it
	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
	CheckContainerStatus(c, cname, "paused")

	// unpause it
	resp, err = request.Post("/containers/" + cname + "/unpause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
	CheckContainerStatus(c, cname, "running")
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
	defer DelContainerForceMultyTime(c, cname)

	// pause it
	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	// start it
	StartContainerOk(c, cname)

	// pause it
	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// should not pause twice
	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	// unpause it
	resp, err = request.Post("/containers/" + cname + "/unpause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// stop it
	StopContainerOk(c, cname)

	// pause it
	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}
