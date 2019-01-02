package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImagePushSuite is the test suite for image create API.
type APIImagePushSuite struct{}

func init() {
	check.Suite(&APIImagePushSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImagePushSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestPushImageFail tests push a non-exist image should fail.
func (suite *APIImagePushSuite) TestPushImageFail(c *check.C) {
	name := "not_exist_image"
	resp, err := request.Post("/images/" + name + "/push")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}
