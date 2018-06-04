package main

import (
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerStopSuite is the test suite for container stop API.
type APIContainerStopSuite struct{}

func init() {
	check.Suite(&APIContainerStopSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerStopSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestStopOk tests a running container could be stopped.
func (suite *APIContainerStopSuite) TestStopOk(c *check.C) {
	cname := "TestStopOk"

	CreateBusyboxContainerOk(c, cname)
	StartContainerOk(c, cname)

	resp, err := request.Post("/containers/" + cname + "/stop")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	DelContainerForceOk(c, cname)
}

// TestNonExistingContainer tests stop a non-existing container return 404.
func (suite *APIContainerStopSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Post("/containers/" + cname + "/stop")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestStopWait tests waiting before stopping container.
func (suite *APIContainerStopSuite) TestStopWait(c *check.C) {
	cname := "TestStopOk"

	CreateBusyboxContainerOk(c, cname)
	StartContainerOk(c, cname)

	q := url.Values{}
	q.Add("t", "1")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/stop", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	DelContainerForceOk(c, cname)
}

// TestInvalidParam tests using invalid parameter return.
func (suite *APIContainerStopSuite) TestInvalidParam(c *check.C) {
	//TODO
	// 1. invalid timeout value
}

// TestStopPausedContainer tests stop a paused container.
func (suite *APIContainerStopSuite) TestStopPausedContainer(c *check.C) {
	cname := "TestStopPausedContainer"

	CreateBusyboxContainerOk(c, cname)
	StartContainerOk(c, cname)

	// pause the container
	PauseContainerOk(c, cname)

	// stop the container
	resp, err := request.Post("/containers/" + cname + "/stop")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// check the container status
	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
	defer resp.Body.Close()

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)
	c.Assert(string(got.State.Status), check.Equals, "stopped")

	DelContainerForceOk(c, cname)
}
