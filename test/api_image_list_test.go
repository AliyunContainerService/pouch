package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImageListSuite is the test suite for image list API.
type APIImageListSuite struct{}

func init() {
	check.Suite(&APIImageListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageListOk tests listing images is OK.
func (suite *APIImageListSuite) TestImageListOk(c *check.C) {
	path := "/images/json"
	resp, err := request.Get(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)
}
