package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerListSuite is the test suite for container list API.
type APIContainerListSuite struct{}

func init() {
	check.Suite(&APIContainerListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestListOk test api is ok with default parameters.
func (suite *APIContainerListSuite) TestListOk(c *check.C) {
	resp, err := request.Get("/containers/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
}
