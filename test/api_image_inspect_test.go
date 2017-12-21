package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImageInspectSuite is the test suite for image inspect API.
type APIImageInspectSuite struct{}

func init() {
	check.Suite(&APIImageInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageInspectOk tests inspecting images is OK.
func (suite *APIImageInspectSuite) TestImageInspectOk(c *check.C) {
	path := "/images/" + busyboxImage + "/json"
	resp, err := request.Get(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)
}
