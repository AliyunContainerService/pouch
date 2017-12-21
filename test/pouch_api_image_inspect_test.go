package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// PouchAPIImageInspectSuite is the test suite for image inspect API.
type PouchAPIImageInspectSuite struct{}

func init() {
	check.Suite(&PouchAPIImageInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIImageInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageInspectOk tests inspecting images is OK.
func (suite *PouchAPIImageInspectSuite) TestImageInspectOk(c *check.C) {
	path := "/images/" + busyboxImage + "/json"
	resp, err := request.Get(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)
}
