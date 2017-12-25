package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerDeleteSuite is the test suite for container delete API.
type APIContainerDeleteSuite struct{}

func init() {
	check.Suite(&APIContainerDeleteSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerDeleteSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestDeleteNonExisting tests deleting a non-existing container return error.
func (suite *APIContainerDeleteSuite) TestDeleteNonExisting(c *check.C) {
	cname := "TestDeleteNonExisting"

	resp, err := request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 404)
}

// TestDeleteRunningCon test deleting running container return 500.
func (suite *APIContainerDeleteSuite) TestDeleteRunningCon(c *check.C) {
	cname := "TestDeleteRunningCon"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	resp, err := request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 500)

	DelContainerForceOk(c, cname)
}

// TestDeletePausedCon test deleting paused container return 500.
func (suite *APIContainerDeleteSuite) TestDeletePausedCon(c *check.C) {
	cname := "TestDeletePausedCon"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	PauseContainerOk(c, cname)

	resp, err := request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 500)

	DelContainerForceOk(c, cname)
}

// TestDeleteStoppedCon test deleting stopped container return 204.
func (suite *APIContainerDeleteSuite) TestDeleteStoppedCon(c *check.C) {
	cname := "TestDeleteStoppedCon"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	StopContainerOk(c, cname)

	resp, err := request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 204)
}

// TestDeleteCreatedCon test deleting created container return 204.
func (suite *APIContainerDeleteSuite) TestDeleteCreatedCon(c *check.C) {
	cname := "TestDeleteCreatedCon"

	CreateBusyboxContainerOk(c, cname)

	resp, err := request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 204)
}
