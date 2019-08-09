package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net"
	"strings"
	"time"

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

// TestContainerExecDetach tests start exec process with detach
func (suite *APIContainerExecStartSuite) TestContainerExecDetach(c *check.C) {
	cname := "TestContainerExecDetach"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)
	StartContainerOk(c, cname)

	// verify that we can obtain the exec process's exitCode = 1
	obj := map[string]interface{}{
		"Cmd":    []string{"sleep", "2", "&&", "exit", "1"},
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
		resp, _, _, err := StartContainerExec(c, execid, false, true)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 200)
	}

	// sleep 2s to wait the process exit
	time.Sleep(2 * time.Second)

	// inspect the exec after exec start
	resp, err = request.Get("/exec/" + execid + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var execInspect02 types.ContainerExecInspect
	request.DecodeBody(&execInspect02, resp.Body)

	c.Assert(execInspect02.Running, check.Equals, false)
	c.Assert(execInspect02.ExitCode, check.Equals, int64(1))

	// verify exitCode = 0
	obj = map[string]interface{}{
		"Cmd":    []string{"ls"},
		"Detach": true,
	}
	body = request.WithJSONBody(obj)
	resp, err = request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	var execCreateResp01 types.ExecCreateResp
	err = request.DecodeBody(&execCreateResp01, resp.Body)
	c.Assert(err, check.IsNil)

	execid = execCreateResp01.ID

	// inspect the exec before exec start
	resp, err = request.Get("/exec/" + execid + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var execInspect03 types.ContainerExecInspect
	request.DecodeBody(&execInspect03, resp.Body)

	c.Assert(execInspect03.Running, check.Equals, false)
	c.Assert(execInspect03.ExitCode, check.Equals, int64(0))

	// start the exec
	{
		resp, _, _, err := StartContainerExec(c, execid, false, true)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 200)
	}

	// inspect the exec after exec start
	resp, err = request.Get("/exec/" + execid + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var execInspect04 types.ContainerExecInspect
	request.DecodeBody(&execInspect04, resp.Body)

	c.Assert(execInspect04.Running, check.Equals, false)
	c.Assert(execInspect04.ExitCode, check.Equals, int64(0))
}
