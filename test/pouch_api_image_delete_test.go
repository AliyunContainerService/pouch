package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// PouchAPIImageDeleteSuite is the test suite for image delete API.
type PouchAPIImageDeleteSuite struct{}

func init() {
	check.Suite(&PouchAPIImageDeleteSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIImageDeleteSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestDeleteNonExisting tests deleting a non-existing image return error.
func (suite *PouchAPIImageDeleteSuite) TestDeleteNonExisting(c *check.C) {
	img := "TestDeleteNonExisting"
	resp, err := request.Delete("/images/" + img)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)
}
