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
}

// TestContainerExecOk tests execing containers is OK.
func (suite *APIContainerExecSuite) TestContainerExecOk(c *check.C) {
	// TODO:
	cname := "TestContainerExecOk"

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
	// start exec process.
	//header := request.WithHeader("Content-Type", "text/plain")
	//resp, err = request.Post("/exec/"+got+"/start", header)
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 201)

	DelContainerForceOk(c, cname)
}
