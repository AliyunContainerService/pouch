package main

import (
	"fmt"
	"reflect"

	"github.com/alibaba/pouch/apis/types"
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
	PullImage(c, fmt.Sprintf("%s:%s", environment.BusyboxRepo, environment.BusyboxTag))
}

// TestImageInspectOk tests inspecting images is OK.
func (suite *APIImageInspectSuite) TestImageInspectOk(c *check.C) {
	var (
		repo   = environment.BusyboxRepo
		tag    = environment.BusyboxTag
		digest = environment.BusyboxDigest
		id     = environment.BusyboxID
	)

	repoTag := fmt.Sprintf("%s:%s", repo, tag)
	repoDigest := fmt.Sprintf("%s@%s", repo, digest)

	for _, image := range []string{
		id,
		repoTag,
		repoDigest,
		fmt.Sprintf("%s:whatever@%s", repo, digest),
	} {
		resp, err := request.Get("/images/" + image + "/json")
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 200)

		got := types.ImageInfo{}
		err = request.DecodeBody(&got, resp.Body)
		c.Assert(err, check.IsNil)

		// TODO: More specific check is needed
		c.Assert(got.Config, check.NotNil)
		c.Assert(got.ID, check.Equals, id)
		c.Assert(got.CreatedAt, check.NotNil)
		c.Assert(got.Size, check.NotNil)
		c.Assert(reflect.DeepEqual(got.RepoTags, []string{repoTag}), check.Equals, true)
		c.Assert(reflect.DeepEqual(got.RepoDigests, []string{repoDigest}), check.Equals, true)
	}
}

// TestImageInspectNotFound tests inspecting non-existing images.
func (suite *APIImageInspectSuite) TestImageInspectNotFound(c *check.C) {
	resp, err := request.Get("/images/" + "TestImageInspectNotFound" + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}
