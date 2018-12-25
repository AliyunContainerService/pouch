package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

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
	cname := "TestContainerExecInspectOk"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	// create an exec and get the execID
	// and inspect the exec before exec start

	execid := CreateExecCmdOk(c, cname, "sleep", "12345678")
	{
		execInspectResp := InspectExecOk(c, execid)
		c.Assert(execInspectResp.Running, check.Equals, false)
		c.Assert(execInspectResp.ExitCode, check.Equals, int64(0))
	}

	// set the detach to be true
	// and start the exec
	{
		obj := map[string]interface{}{
			"Detach": true,
		}
		body := request.WithJSONBody(obj)
		resp, err := request.Post(fmt.Sprintf("/exec/%s/start", execid), body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 200)
	}

	// inspect the exec after exec start
	{
		execInspectResp := InspectExecOk(c, execid)
		c.Assert(execInspectResp.Running, check.Equals, true)
		c.Assert(execInspectResp.ExitCode, check.Equals, int64(0))
	}

	// find the exec command and terminate it.
	{
		cmd := exec.Command("bash", "-c", "ps aux | grep 'sleep 12345678' | awk '{print $2}'")
		output, err := cmd.Output()
		c.Assert(err, check.IsNil)

		outputStr := strings.TrimSpace(string(output))
		pids := strings.SplitN(outputStr, "\n", -1)
		c.Assert(len(pids), check.Equals, 1)

		// kill the exec process by sending terminal signal
		pid, err := strconv.Atoi(pids[0])
		c.Assert(err, check.IsNil)
		err = syscall.Kill(pid, syscall.SIGTERM)
		c.Assert(err, check.IsNil)

		execInspectResp := InspectExecOk(c, execid)
		c.Assert(execInspectResp.Running, check.Equals, false)
		c.Assert(execInspectResp.ExitCode, check.Equals, int64(0))
	}
}
