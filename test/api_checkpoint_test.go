package main

import (
	"fmt"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerCheckpointSuite is the test suite for container create API.
type APIContainerCheckpointSuite struct{}

func init() {
	check.Suite(&APIContainerCheckpointSuite{})
	fmt.Println("hello travis-ci")
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerCheckpointSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestCheckpointCreateAPINonExist tests create a non-existing container checkpoint return error.
func (suite *APIContainerCheckpointSuite) TestCheckpointCreateAPINonExist(c *check.C) {
	cname := "TestCheckpointCreateAPINonExist"

	body := request.WithJSONBody(map[string]interface{}{})
	resp, err := request.Post("/containers/"+cname+"/checkpoints", body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 400)
}

// TestCheckpointCreateAPI tests checkpoint create api
func (suite *APIContainerCheckpointSuite) TestCheckpointCreateAPI(c *check.C) {
	SkipIfFalse(c, environment.IsCRIUExist)
	cname := "TestCheckpointCreateAPI"

	defer DelContainerForceMultyTime(c, cname)
	CreateBusyboxContainerOk(c, cname)
	StartContainerOk(c, cname)

	obj := map[string]interface{}{
		"CheckpointID": "cp0",
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/checkpoints", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	CheckContainerRunning(c, cname, true)

	// create checkpoint leaving conainer exited
	obj = map[string]interface{}{
		"CheckpointID": "cp1",
		"Exit":         true,
	}

	body = request.WithJSONBody(obj)
	resp, err = request.Post("/containers/"+cname+"/checkpoints", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	CheckContainerRunning(c, cname, false)
}

// TestCheckpointListAPINonExist tests list a non-existing container checkpoint return error.
func (suite *APIContainerCheckpointSuite) TestCheckpointListAPINonExist(c *check.C) {
	cname := "TestCheckpointListAPINonExist"

	body := request.WithJSONBody(map[string]interface{}{})
	resp, err := request.Get("/containers/"+cname+"/checkpoints", body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 404)
}

// TestCheckpointListAPI tests checkpoint list api
func (suite *APIContainerCheckpointSuite) TestCheckpointListAPI(c *check.C) {
	SkipIfFalse(c, environment.IsCRIUExist)
	cname := "TestCheckpointListAPI"

	defer DelContainerForceMultyTime(c, cname)
	CreateBusyboxContainerOk(c, cname)
	StartContainerOk(c, cname)

	resp, err := request.Get("/containers/" + cname + "/checkpoints")
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 200)

	// Decode response
	c.Assert(err, check.IsNil)

}

// TestCheckpointDelAPINonExist tests delete a non-existing container checkpoint return error.
func (suite *APIContainerCheckpointSuite) TestCheckpointDelAPINonExist(c *check.C) {
	cname := "TestCheckpointDelAPINonExist"

	resp, err := request.Delete("/containers/" + cname + "/checkpoints/" + cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 404)
}

// TestCheckpointDelAPI tests checkpoint delete api
func (suite *APIContainerCheckpointSuite) TestCheckpointDelAPI(c *check.C) {
	SkipIfFalse(c, environment.IsCRIUExist)
	cname := "TestCheckpointDelAPI"

	defer DelContainerForceMultyTime(c, cname)
	CreateBusyboxContainerOk(c, cname)
	StartContainerOk(c, cname)

	obj := map[string]interface{}{
		"CheckpointID": "cp0",
	}
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/checkpoints", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	resp, err = request.Delete("/containers/" + cname + "/checkpoints/" + "cp0")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
}
