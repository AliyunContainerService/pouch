package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerStartSuite is the test suite for container pause/unpause API.
type APIContainerStartSuite struct{}

func init() {
	check.Suite(&APIContainerStartSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerStartSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestStartOk tests starting container could work.
func (suite *APIContainerStartSuite) TestStartOk(c *check.C) {
	cname := "TestStartOk"

	CreateBusyboxContainerOk(c, cname)

	resp, err := request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	DelContainerForceOk(c, cname)
}

// TestNonExistingContainer tests start a non-existing container return 404.
func (suite *APIContainerStartSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"

	resp, err := request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestStartStoppedContainer tests start a contain in stopped state is OK.
func (suite *APIContainerStartSuite) TestStartStoppedContainer(c *check.C) {
	cname := "TestStartStoppedContainer"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	StopContainerOk(c, cname)

	resp, err := request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	DelContainerForceOk(c, cname)
}

// TestStartPausedContainer tests start a contain in paused state will fail.
func (suite *APIContainerStartSuite) TestStartPausedContainer(c *check.C) {
	cname := "TestStartPausedContainer"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	PauseContainerOk(c, cname)

	resp, err := request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	DelContainerForceOk(c, cname)
}

// TestStartDetachKeyWork test detatch-keys works.
func (suite *APIContainerStartSuite) TestStartDetachKeyWork(c *check.C) {
	cname := "TestStartDetachKeyWork"

	CreateBusyboxContainerOk(c, cname)

	q := url.Values{}
	q.Add("detachKeys", "EOF")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/start", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// TODO: check the "EOF" detatchkey really works.

	DelContainerForceOk(c, cname)
}

// TestInvalidParam tests using invalid parameter return.
func (suite *APIContainerStartSuite) TestInvalidParam(c *check.C) {
	//TODO
}
