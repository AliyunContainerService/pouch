package main

import (
	"github.com/alibaba/pouch/apis/types"
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

	StartContainerOk(c, cname)

	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// add state check
	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)
	defer resp.Body.Close()

	c.Assert(string(got.State.Status), check.Equals, "paused")

	resp, err = request.Post("/containers/" + cname + "/unpause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// add state check
	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got = types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)
	defer resp.Body.Close()

	c.Assert(string(got.State.Status), check.Equals, "running")

	DelContainerForceMultyTime(c, cname)
}

// TestPauseNonExistingContainer tests pause a non-existing container return 404.
func (suite *APIContainerPauseSuite) TestPauseNonExistingContainer(c *check.C) {
	cname := "TestPauseNonExistingContainer"
	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestPauseUnpauseNotRunningContainer tests pausing/unpausing a non-running container will return error.
func (suite *APIContainerPauseSuite) TestPauseUnpauseNotRunningContainer(c *check.C) {
	cname := "TestPauseUnpauseNotRunningContainer"
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

	DelContainerForceMultyTime(c, cname)
}
