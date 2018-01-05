package main

import (
	"bufio"
	"io"
	"net"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerExecStartSuite is the test suite for container exec start API.
type APIContainerExecStartSuite struct{}

func init() {
	check.Suite(&APIContainerExecStartSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerExecStartSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

func checkEchoSuccess(c *check.C, conn net.Conn, br *bufio.Reader, exp string) {
	defer conn.Close()

	// Allocate a large space incase there is error.
	got := make([]byte, len(exp))
	_, err := io.ReadFull(br, got)
	c.Assert(err, check.IsNil)
	c.Assert(string(got), check.Equals, exp, check.Commentf("Expected %s, got %s", exp, string(got)))
}

// TestContainerExecStartWithoutUpgrade tests start exec without upgrade which will return 200 OK.
func (suite *APIContainerExecStartSuite) TestContainerExecStartWithoutUpgrade(c *check.C) {
	cname := "TestContainerCreateExecStartWithoutUpgrade"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	execid := CreateExecEchoOk(c, cname)

	obj := map[string]interface{}{}
	resp, conn, br, err := request.Hijack("/exec/"+execid+"/start", request.WithJSONBody(obj))
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	checkEchoSuccess(c, conn, br, "test")

	DelContainerForceOk(c, cname)
}

// TestContainerExecStartOk tests start exec.
func (suite *APIContainerExecStartSuite) TestContainerExecStart(c *check.C) {
	cname := "TestContainerCreateExecStart"

	CreateBusyboxContainerOk(c, cname)

	StartContainerOk(c, cname)

	execid := CreateExecEchoOk(c, cname)

	resp, conn, reader, err := StartContainerExec(c, execid, false, false)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 101)

	checkEchoSuccess(c, conn, reader, "test")

	DelContainerForceOk(c, cname)
}

// TestContainerExecStartNotFound tests starting an non-existing execID return error.
func (suite *APIContainerExecStartSuite) TestContainerExecStartNotFound(c *check.C) {
	obj := map[string]interface{}{
		"Detach": false,
		"Tty":    false,
	}
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/exec/TestContainerExecStartNotFound/start", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestContainerExecStartStopped tests start a process in a stopped container return error.
func (suite *APIContainerExecStartSuite) TestContainerExecStartStopped(c *check.C) {
}

// TestContainerExecStartPaused tests start a process in a paused container return error.
func (suite *APIContainerExecStartSuite) TestContainerExecStartPaused(c *check.C) {
}

// TestContainerExecStartDup tests start a process twice return error.
func (suite *APIContainerExecStartSuite) TestContainerExecStartDup(c *check.C) {
}
