package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIContainerExecSuite is the test suite for container exec API.
type APIContainerExecSuite struct{}

func init() {
	check.Suite(&APIContainerExecSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerExecSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
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

	var got string
	request.DecodeBody(&got, resp.Body)
	c.Logf("ExecID is %s", got)

	DelContainerForceOk(c, cname)
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

	DelContainerForceOk(c, cname)
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

	DelContainerForceOk(c, cname)
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

	DelContainerForceOk(c, cname)
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

	DelContainerForceOk(c, cname)
}
