package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/docker/docker/pkg/stdcopy"
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

	PullImage(c, busyboxImage)
}

func checkEchoSuccess(c *check.C, tty bool, conn net.Conn, br *bufio.Reader, exp string) {
	defer conn.Close()

	// Allocate a large space incase there is error.
	var (
		buf bytes.Buffer
		err error
	)

	if !tty {
		_, err = stdcopy.StdCopy(&buf, &buf, br)
	} else {
		_, err = io.Copy(&buf, br)
	}
	c.Assert(err, check.IsNil)
	c.Assert(strings.TrimSpace(buf.String()), check.Equals, exp, check.Commentf("Expected %s, got %s", exp, buf.String()))
}

func (suite *APIContainerExecStartSuite) TestContainerExecWithLongInput(c *check.C) {
	cname := "TestContainerExecWithLongInput"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)
	StartContainerOk(c, cname)

	var execID string

	// create exec
	{
		// echo the input
		obj := map[string]interface{}{
			"Cmd":          []string{"cat"},
			"AttachStderr": true,
			"AttachStdout": true,
			"AttachStdin":  true,
			"Tty":          true, // avoid to use docker stdcopy
		}

		body := request.WithJSONBody(obj)
		resp, err := request.Post("/containers/"+cname+"/exec", body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 201)

		var got types.ExecCreateResp
		c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
		execID = got.ID
	}

	// start exec
	{
		obj := map[string]interface{}{
			"Tty": true, // avoid to use docker stdcopy
		}
		resp, conn, br, err := request.Hijack("/exec/"+execID+"/start", request.WithJSONBody(obj))
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 200)
		defer conn.Close()

		// input 256 chars
		input := make([]byte, 256)
		_, err = rand.Read(input)
		c.Assert(err, check.IsNil)
		content := hex.EncodeToString(input)

		_, err = conn.Write([]byte(content + "\n"))
		c.Assert(err, check.IsNil)

		reader := bufio.NewReader(br)
		got, _, err := reader.ReadLine()
		c.Assert(err, check.IsNil)
		c.Assert(string(got), check.Equals, content)
	}
}

// TestContainerExecStartWithoutUpgrade tests start exec without upgrade which will return 200 OK.
func (suite *APIContainerExecStartSuite) TestContainerExecStartWithoutUpgrade(c *check.C) {
	cname := "TestContainerCreateExecStartWithoutUpgrade"
	content := "hi pouch"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)
	execid := CreateExecEchoOk(c, cname, content)

	obj := map[string]interface{}{}
	resp, conn, br, err := request.Hijack("/exec/"+execid+"/start", request.WithJSONBody(obj))
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
	checkEchoSuccess(c, false, conn, br, content)
}

// TestContainerExecStartOk tests start exec.
func (suite *APIContainerExecStartSuite) TestContainerExecStart(c *check.C) {
	cname := "TestContainerCreateExecStart"
	content := "test"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	execid := CreateExecEchoOk(c, cname, content)

	resp, conn, reader, err := StartContainerExec(c, execid, false, false)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 101)
	checkEchoSuccess(c, false, conn, reader, content)
}

// TestContainerExecStartNotFound tests starting an non-existing execID return error.
func (suite *APIContainerExecStartSuite) TestContainerExecStartNotFound(c *check.C) {
	obj := map[string]interface{}{}
	body := request.WithJSONBody(obj)

	resp, err := request.Post("/exec/TestContainerExecStartNotFound/start", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestContainerExecStartStopped tests start a process in a stopped container return error.
func (suite *APIContainerExecStartSuite) TestContainerExecStartStopped(c *check.C) {
	// TODO: missing case
	helpwantedForMissingCase(c, "container api exec start stoped case")
}

// TestContainerExecStartPaused tests start a process in a paused container return error.
func (suite *APIContainerExecStartSuite) TestContainerExecStartPaused(c *check.C) {
	// TODO: missing case
	helpwantedForMissingCase(c, "container api exec start paused case")
}

// TestContainerExecStartDup tests start a process twice return error.
func (suite *APIContainerExecStartSuite) TestContainerExecStartDup(c *check.C) {
	// TODO: missing case
	helpwantedForMissingCase(c, "container api exec start twice case")
}
