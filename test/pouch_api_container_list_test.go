package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// PouchAPIContainerListSuite is the test suite for container list API.
type PouchAPIContainerListSuite struct{}

func init() {
	check.Suite(&PouchAPIContainerListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIContainerListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestListOk test api is ok with default parameters.
func (suite *PouchAPIContainerListSuite) TestListOk(c *check.C) {
	resp, err := request.Get("/containers/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)
}
