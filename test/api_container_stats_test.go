package main

import (
	"fmt"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIContainerStatsSuite is the test suite for container stats API.
type APIContainerStatsSuite struct{}

func init() {
	check.Suite(&APIContainerStatsSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerStatsSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage125)
}

// TestNoSuchContainer tests a container that doesn't exits return error.
func (suite *APIContainerStatsSuite) TestNoSuchContainer(c *check.C) {
	resp, err := request.Get("/containers/nosuchcontainer/stats")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, http.StatusNotFound)
}

// TestNoStream tests stats api without stream
func (suite *APIContainerStatsSuite) TestNoStream(c *check.C) {
	name := "test_no_stream"
	command.PouchRun("run", "-d", "--name", name, "--net", "none", busyboxImage, "sh", "-c", "while true; do sleep 1; done").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	resp, err := request.Get(fmt.Sprintf("/containers/%s/stats", name))
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, http.StatusOK)

	out := types.ContainerStats{}
	err = request.DecodeBody(&out, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(out.Name, check.Equals, name)
	c.Assert(out.ID, check.NotNil)
	c.Assert(out.Read, check.NotNil)
	c.Assert(out.CPUStats, check.NotNil)
	c.Assert(out.MemoryStats, check.NotNil)
	c.Assert(out.BlkioStats, check.NotNil)
	c.Assert(out.PidsStats, check.NotNil)
	// test net=none container that should not return network info
	c.Assert(out.Networks, check.IsNil)
}
