package main

import (
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerExecInspectSuite is the test suite for container exec inspect API.
type APIContainerExecInspectSuite struct{}

func init() {
	check.Suite(&APIContainerExecInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerExecInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestContainerCreateExecOk tests execing containers is OK.
func (suite *APIContainerExecInspectSuite) TestContainerExecInspectOk(c *check.C) {
	c.Skip("skip flaky test due to issue#1372")
	cname := "TestContainerExecInspectOk"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	obj := map[string]interface{}{
		"Cmd":    []string{"sleep", "9"},
		"Detach": true,
	}
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	var execCreateResp types.ExecCreateResp
	err = request.DecodeBody(&execCreateResp, resp.Body)
	c.Assert(err, check.IsNil)

	execid := execCreateResp.ID

	// inspect the exec before exec start
	resp, err = request.Get("/exec/" + execid + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var execInspect01 types.ContainerExecInspect
	request.DecodeBody(&execInspect01, resp.Body)

	c.Assert(execInspect01.Running, check.Equals, false)
	c.Assert(execInspect01.ExitCode, check.Equals, int64(0))

	// start the exec
	{
		resp, conn, _, err := StartContainerExec(c, execid, false, false)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 101)
		c.Assert(conn.Close(), check.IsNil)
	}

	// inspect the exec after exec start
	resp, err = request.Get("/exec/" + execid + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var execInspect02 types.ContainerExecInspect
	request.DecodeBody(&execInspect02, resp.Body)

	c.Assert(execInspect02.Running, check.Equals, true)
	c.Assert(execInspect02.ExitCode, check.Equals, int64(0))

	// sleep 10s to wait the process exit
	time.Sleep(10 * time.Second)

	resp, err = request.Get("/exec/" + execid + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var execInspect03 types.ContainerExecInspect
	request.DecodeBody(&execInspect03, resp.Body)

	c.Assert(execInspect03.Running, check.Equals, false)
	c.Assert(execInspect03.ExitCode, check.Equals, int64(0))

	DelContainerForceOk(c, cname)
}
