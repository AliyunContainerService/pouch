package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerDeleteSuite is the test suite for container delete API.
type APIContainerDeleteSuite struct{}

func init() {
	check.Suite(&APIContainerDeleteSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerDeleteSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestDeleteNonExisting tests deleting a non-existing container return error.
func (suite *APIContainerDeleteSuite) TestDeleteNonExisting(c *check.C) {
	cname := "TestDeleteNonExisting"
	resp, err := request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)
}
