package main

import (
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
	defer DelContainerForceMultyTime(c, cname)
	StartContainerOk(c, cname)
}

// TestNonExistingContainer tests start a non-existing container return 404.
func (suite *APIContainerStartSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"

	body := request.WithJSONBody(map[string]interface{}{})
	resp, err := request.Post("/containers/"+cname+"/start", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestStartStoppedContainer tests start a contain in stopped state is OK.
func (suite *APIContainerStartSuite) TestStartStoppedContainer(c *check.C) {
	cname := "TestStartStoppedContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)
	StartContainerOk(c, cname)
	StopContainerOk(c, cname)
	StartContainerOk(c, cname)
}

// TestStartPausedContainer tests start a contain in paused state will fail.
func (suite *APIContainerStartSuite) TestStartPausedContainer(c *check.C) {
	cname := "TestStartPausedContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)
	PauseContainerOk(c, cname)

	body := request.WithJSONBody(map[string]interface{}{})
	resp, err := request.Post("/containers/"+cname+"/start", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 409)
}

// TestStartDetachKeyWork test detatch-keys works.
func (suite *APIContainerStartSuite) TestStartDetachKeyWork(c *check.C) {
	cname := "TestStartDetachKeyWork"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	body := request.WithJSONBody(map[string]interface{}{
		"DetachKeys": "EOF",
	})
	resp, err := request.Post("/containers/"+cname+"/start", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	// TODO: missing case
	//
	//	check the "EOF" detatchkey really works.
}

// TestInvalidParam tests using invalid parameter return.
func (suite *APIContainerStartSuite) TestInvalidParam(c *check.C) {
	// TODO: missing case
	helpwantedForMissingCase(c, "container api start bad request")
}
