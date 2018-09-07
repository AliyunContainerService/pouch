package main

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"strings"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/go-check/check"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"
)

// APIContainerUpdateSuite is the test suite for container attach API.
type APIContainerUpdateSuite struct{}

func init() {
	check.Suite(&APIContainerUpdateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerUpdateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

func checkEqual(c *check.C, tty bool, conn net.Conn, br *bufio.Reader, exp string) {
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

// TestStoppedContainerUpdateHostname test stopped container update hostname
func (suite *APIContainerUpdateSuite) TestStoppedContainerUpdateHostname(c *check.C) {
	cname := "TestUpdateOk"
	newHostname := "abcdefgh"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)
	StopContainerOk(c, cname)

	UpdateContaineHostnamerOK(c, cname, newHostname)
	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(string(got.Config.Hostname), check.Equals, newHostname)
}

// TestStoppedContainerUpdateHostname test stopped container update hostname
func (suite *APIContainerUpdateSuite) TestRunningContainerUpdateHostname(c *check.C) {
	cname := "TestUpdateOk"
	newHostname := "abcdefgh"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	UpdateContaineHostnamerOK(c, cname, newHostname)
	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(string(got.Config.Hostname), check.Equals, newHostname)

	execID := CreateExecHostnameOk(c, cname)
	obj := map[string]interface{}{}
	resp, conn, br, err := request.Hijack("/exec/"+execID+"/start", request.WithJSONBody(obj))
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
	checkEqual(c, false, conn, br, newHostname)
}
