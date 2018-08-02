package main

import (
	"time"

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

	PullImage(c, busyboxImage)
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
	defer DelContainerForceMultyTime(c, cname)

	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	// TODO: missing case
	//
	// add more field checker
	c.Assert(got.Image, check.Equals, busyboxImage)
	c.Assert(got.Name, check.Equals, cname)
	c.Assert(got.Created, check.NotNil)
	// StartedAt time should be 0001-01-01T00:00:00Z for a non-started container
	c.Assert(got.State.StartedAt, check.Equals, time.Time{}.UTC().Format(time.RFC3339Nano))
	// FinishAt time should be 0001-01-01T00:00:00Z for a non-stopped container
	c.Assert(got.State.FinishedAt, check.Equals, time.Time{}.UTC().Format(time.RFC3339Nano))
}

// TestInspectPid tests the response of inspect has process pid.
func (suite *APIContainerInspectSuite) TestInspectPid(c *check.C) {
	cname := "TestInspectPid"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
	c.Assert(got.State.Pid, check.NotNil)
}
