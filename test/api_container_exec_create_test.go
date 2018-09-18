package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerExecSuite is the test suite for container exec API.
type APIContainerExecSuite struct{}

func init() {
	check.Suite(&APIContainerExecSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerExecSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestContainerCreateExecOk tests execing containers is OK.
func (suite *APIContainerExecSuite) TestContainerCreateExecOk(c *check.C) {
	cname := "TestContainerCreateExecOk"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	obj := map[string]interface{}{
		"Cmd":    []string{"echo", "test"},
		"Detach": true,
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)
}

// TestContainerCreateExecWithEnvs tests execing containers with Envs.
func (suite *APIContainerExecSuite) TestContainerCreateExecWithEnvs(c *check.C) {
	cname := "TestContainerCreateExecWithEnvs"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	obj := map[string]interface{}{
		"Cmd":    []string{"Env"},
		"Detach": true,
		"Env":    []string{"Test=OK"},
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)
}

// TestContainerCreateExecNoCmd tests execing containers is OK.
func (suite *APIContainerExecSuite) TestContainerCreateExecNoCmd(c *check.C) {
	cname := "TestContainerCreateExecNoCmd"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	obj := map[string]interface{}{
		"Detach": true,
	}
	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)
}

// TestExecCreatedContainer tests creating exec on created container return error.
func (suite *APIContainerExecSuite) TestExecCreatedContainer(c *check.C) {
	cname := "TestExecCreatedContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	obj := map[string]interface{}{
		"Cmd":    []string{"echo", "test"},
		"Detach": true,
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}

// TestExecPausedContainer tests creating exec on paused container return error.
func (suite *APIContainerExecSuite) TestExecPausedContainer(c *check.C) {
	cname := "TestExecPausedContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)
	PauseContainerOk(c, cname)

	obj := map[string]interface{}{
		"Cmd":    []string{"echo", "test"},
		"Detach": true,
	}
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}

// TestExecStoppedContainer tests creating exec on stopped container return error.
func (suite *APIContainerExecSuite) TestExecStoppedContainer(c *check.C) {
	cname := "TestExecStoppedContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)
	StopContainerOk(c, cname)

	obj := map[string]interface{}{
		"Cmd":    []string{"echo", "test"},
		"Detach": true,
	}
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}
