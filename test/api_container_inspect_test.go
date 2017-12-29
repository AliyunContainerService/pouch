package main

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerInspectSuite is the test suite for container inspect API.
type APIContainerInspectSuite struct{}

func init() {
	check.Suite(&APIContainerInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestInspectNoSuchContainer tests inspecting a container that doesn't exits return error.
func (suite *APIContainerInspectSuite) TestInspectNoSuchContainer(c *check.C) {
	resp, err := request.Get("/containers/nosuchcontainerxxx/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestInspectOk tests inspecting an existing container is OK.
func (suite *APIContainerInspectSuite) TestInpectOk(c *check.C) {
	cname := "TestInpectOk"

	CreateBusyboxContainerOk(c, cname)

	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(got.Image, check.Equals, busyboxImage)
	c.Assert(got.Name, check.Equals, cname)
	c.Assert(got.Created, check.NotNil)

	DelContainerForceOk(c, cname)
}

// TestNonExistingContainer tests inspect a non-existing container return 404.
func (suite *APIContainerInspectSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestInspectPid tests the response of inspect has process pid.
func (suite *APIContainerInspectSuite) TestInspectPid(c *check.C) {
	cname := "TestInspectPid"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(got.State.Pid, check.NotNil)

	DelContainerForceOk(c, cname)
}
