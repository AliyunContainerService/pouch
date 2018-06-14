package main

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"

	"net/url"
)

// APIContainerKillSuite is the test suite for container kill API.
type APIContainerKillSuite struct{}

func init() {
	check.Suite(&APIContainerKillSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerKillSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestKillOk tests a running container could be killed.
func (suite *APIContainerKillSuite) TestKillOk(c *check.C) {
	cname := "TestKillOk"
	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	q := url.Values{}
	q.Add("signal", "SIGKILL")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/kill", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// add state check here
	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
	defer resp.Body.Close()

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(string(got.State.Status), check.Equals, "exited")

	DelContainerForceMultyTime(c, cname)
}

// TestKillNonExistingContainer tests killing a non-existing container return 404.
func (suite *APIContainerKillSuite) TestKillNonExistingContainer(c *check.C) {
	cname := "TestKillNonExistingContainer"

	q := url.Values{}
	q.Add("signal", "SIGKILL")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/kill", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestKillNotRunningContainer tests killing a non-running container will return error.
func (suite *APIContainerKillSuite) TestKillNotRunningContainer(c *check.C) {
	cname := "TestKillNotRunningContainer"
	CreateBusyboxContainerOk(c, cname)

	q := url.Values{}
	q.Add("signal", "SIGKILL")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/kill", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	StartContainerOk(c, cname)

	resp, err = request.Post("/containers/"+cname+"/kill", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	DelContainerForceMultyTime(c, cname)
}

// TestKillContainerWithInvalidSignal tests killing a container with invalid signal will return error.
func (suite *APIContainerKillSuite) TestKillContainerWithInvalidSignal(c *check.C) {
	cname := "TestKillContainerWithInvalidSignal"
	CreateBusyboxContainerOk(c, cname)

	q := url.Values{}
	q.Add("signal", "0")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/kill", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 400)

	StartContainerOk(c, cname)

	q = url.Values{}
	q.Add("signal", "SIGKILL")
	query = request.WithQuery(q)

	resp, err = request.Post("/containers/"+cname+"/kill", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	DelContainerForceMultyTime(c, cname)
}
