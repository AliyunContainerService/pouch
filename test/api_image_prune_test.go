package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIImagePruneSuite is the test suite for image prune API.
type APIImagePruneSuite struct{}

func init() {
	check.Suite(&APIImagePruneSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImagePruneSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestPruneUnusedImageOk tests delete unused image is ok.
func (suite *APIImagePruneSuite) TestPruneUnusedImageOk(c *check.C) {
	PullImage(c, helloworldImage)
	resp, err := request.Post("/images/prune")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	resp, err = request.Get("/images/" + helloworldImage + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestPruneUsedImageOk tests cannot delete used image is ok.
func (suite *APIImagePruneSuite) TestPruneUsedImageOk(c *check.C) {
	containerA := "ContainerA"
	command.PouchRun("run", "-d", "--name", containerA, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerA)

	containerB := "ContainerB"
	command.PouchRun("create", "--name", containerB, "-l", "label="+containerB, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerB)

	resp, err := request.Post("/images/prune")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	resp, err = request.Get("/images/" + busyboxImage125 + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
}
