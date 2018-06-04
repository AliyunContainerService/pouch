package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImageDeleteSuite is the test suite for image delete API.
type APIImageDeleteSuite struct{}

func init() {
	check.Suite(&APIImageDeleteSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageDeleteSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestDeleteNonExisting tests deleting a non-existing image return error.
func (suite *APIImageDeleteSuite) TestDeleteNonExisting(c *check.C) {
	img := "TestDeleteNonExisting"
	resp, err := request.Delete("/images/" + img)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestDeleteUsingImage tests deleting an image in use by running container will fail.
func (suite *APIImageDeleteSuite) TestDeleteUsingImage(c *check.C) {
	cname := "TestDeleteUsingImage"
	CreateBusyboxContainerOk(c, cname)
	StartContainerOk(c, cname)

	resp, err := request.Delete("/images/" + busyboxImage)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)

	DelContainerForceOk(c, cname)
}
