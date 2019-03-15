package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerRestartSuite is the test suite for container upgrade API.
type APIContainerRestartSuite struct{}

func init() {
	check.Suite(&APIContainerRestartSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerRestartSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestAPIContainerRestart is to verify restarting container.
func (suite *APIContainerRestartSuite) TestAPIContainerRestart(c *check.C) {
	cname := "TestAPIContainerRestart"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	q := url.Values{}
	q.Add("t", "1")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/restart", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
}

// TestAPIRestartStoppedContainer it to verify restarting a stopped container.
func (suite *APIContainerRestartSuite) TestAPIRestartStoppedContainer(c *check.C) {
	cname := "TestAPIRestartStoppedContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	q := url.Values{}
	q.Add("t", "1")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/restart", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
}

// TestAPIRestartPausedContainer is to verify restarting a paused container.
func (suite *APIContainerRestartSuite) TestAPIRestartPausedContainer(c *check.C) {
	cname := "TestAPIRestartPauseContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)
	PauseContainerOk(c, cname)

	q := url.Values{}
	q.Add("t", "1")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/restart", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
}
