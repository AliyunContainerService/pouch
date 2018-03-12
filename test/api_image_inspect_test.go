package main

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIImageInspectSuite is the test suite for image inspect API.
type APIImageInspectSuite struct{}

func init() {
	check.Suite(&APIImageInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TestImageInspectOk tests inspecting images is OK.
func (suite *APIImageInspectSuite) TestImageInspectOk(c *check.C) {
	resp, err := request.Get("/images/" + busyboxImage + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ImageInfo{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	// TODO: More specific check is needed
	c.Assert(got.Config, check.NotNil)
	c.Assert(got.Name, check.Equals, busyboxImage)
	c.Assert(got.ID, check.NotNil)
	c.Assert(got.CreatedAt, check.NotNil)
	c.Assert(got.Digest, check.Matches, "sha256.*")
	c.Assert(got.Size, check.NotNil)
	c.Assert(got.Tag, check.Equals, "latest")
}

// TestImageInspectNotFound tests inspecting non-existing images.
func (suite *APIImageInspectSuite) TestImageInspectNotFound(c *check.C) {
	resp, err := request.Get("/images/" + "TestImageInspectNotFound" + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}
