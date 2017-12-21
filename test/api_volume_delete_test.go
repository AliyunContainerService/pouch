package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIVolumeDeleteSuite is the test suite for volume delete API.
type APIVolumeDeleteSuite struct{}

func init() {
	check.Suite(&APIVolumeDeleteSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIVolumeDeleteSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestDeleteNonExisting tests deleting a non-existing volume return error.
func (suite *APIVolumeDeleteSuite) TestDeleteNonExisting(c *check.C) {
	vol := "TestDeleteNonExisting"
	path := "/volumes/" + vol
	resp, err := request.Delete(path)
	c.Assert(err, check.IsNil)
	// TODO: Now server return 500, should return 404
	c.Assert(resp.StatusCode, check.Equals, 500)
}
