package main

import (
	"github.com/alibaba/pouch/apis/types"
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
	// TODO:
	cname := "TestContainerCreateExecOk"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	obj := map[string]interface{}{
		"Cmd":    []string{"echo", "test"},
		"Detach": true,
	}
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	var execCreateResp types.ExecCreateResp
	request.DecodeBody(&execCreateResp, resp.Body)
	c.Logf("ExecID is %s", execCreateResp.ID)

	DelContainerForceMultyTime(c, cname)
}

// TestContainerCreateExecNoCmd tests execing containers is OK.
func (suite *APIContainerExecSuite) TestContainerCreateExecNoCmd(c *check.C) {
	cname := "TestContainerCreateExecNoCmd"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	// TODO: also test "Cmd":nil
	obj := map[string]interface{}{
		"Cmd":    []string{""},
		"Detach": true,
	}
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	DelContainerForceMultyTime(c, cname)
}

// TestExecCreatedContainer tests creating exec on created container return error.
func (suite *APIContainerExecSuite) TestExecCreatedContainer(c *check.C) {
	// TODO:
	cname := "TestExecCreatedContainer"

	CreateBusyboxContainerOk(c, cname)

	obj := map[string]interface{}{
		"Cmd":    []string{"echo", "test"},
		"Detach": true,
	}
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	DelContainerForceMultyTime(c, cname)
}

// TestExecPausedContainer tests creating exec on paused container return error.
func (suite *APIContainerExecSuite) TestExecPausedContainer(c *check.C) {
	cname := "TestExecPausedContainer"

	CreateBusyboxContainerOk(c, cname)

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

	DelContainerForceMultyTime(c, cname)
}

// TestExecStoppedContainer tests creating exec on stopped container return error.
func (suite *APIContainerExecSuite) TestExecStoppedContainer(c *check.C) {
	cname := "TestExecStoppedContainer"

	CreateBusyboxContainerOk(c, cname)

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

	DelContainerForceMultyTime(c, cname)
}
