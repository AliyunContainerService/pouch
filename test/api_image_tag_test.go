package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIImageTagSuite is the test suite for image tag API.
type APIImageTagSuite struct{}

func init() {
	check.Suite(&APIImageTagSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageTagSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestImageTagCreateWithTagOK tests OK.
func (suite *APIImageTagSuite) TestImageTagCreateWithTagOK(c *check.C) {
	repo, tag := "localhost:5000/testimagetagok/pouch", "0.5.0"
	resp, err := suite.doTag(busyboxImage, repo, tag)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	tagRef := fmt.Sprintf("%s:%s", repo, tag)
	defer DelImageForceOk(c, tagRef)
	suite.checkTagReferenceExisting(c, tagRef, true)
}

// TestImageTagCreateUsingDefaultTagOK tests OK.
func (suite *APIImageTagSuite) TestImageTagCreateUsingDefaultTagOK(c *check.C) {
	repo, tag := "localhost:5000/testimagetagok/pouch", "latest"
	resp, err := suite.doTag(busyboxImage, repo, tag)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	tagRef := fmt.Sprintf("%s:%s", repo, tag)
	defer DelImageForceOk(c, tagRef)
	suite.checkTagReferenceExisting(c, tagRef, true)
}

// TestImageTagUsingNoFoundSourceImage tests fail.
func (suite *APIImageTagSuite) TestImageTagUsingNoFoundSourceImage(c *check.C) {
	repo, tag := "image_test_using_no_found_source_image", "latest"
	resp, err := suite.doTag("image_test_ghost_image", repo, tag)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestImageTagFailToOverrideExistingPrimaryReference tests fail.
func (suite *APIImageTagSuite) TestImageTagFailToOverrideExistingPrimaryReference(c *check.C) {
	repo, tag := environment.BusyboxRepo, environment.Busybox125Tag
	tagRef := fmt.Sprintf("%s:%s", repo, tag)
	command.PouchRun("pull", tagRef).Assert(c, icmd.Success)
	defer DelImageForceOk(c, tagRef)

	resp, err := suite.doTag(busyboxImage, repo, tag)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 400)
}

// TestImageTagFailToOverrideExistingTag tests fail.
func (suite *APIImageTagSuite) TestImageTagFailToOverrideExistingTag(c *check.C) {
	repo, tag := "localhost:5000/shouldnotoverridetag", "1.0"
	tagRef := fmt.Sprintf("%s:%s", repo, tag)

	// create valid tag
	{
		command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)

		resp, err := suite.doTag(busyboxImage, repo, tag)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 201)
		defer DelImageForceOk(c, tagRef)
	}

	// override existing tag
	{
		command.PouchRun("pull", busyboxImage125).Assert(c, icmd.Success)
		defer DelImageForceOk(c, busyboxImage125)

		resp, err := suite.doTag(busyboxImage125, repo, tag)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 400)
	}
}

// TestImageTagFailToUseSha256AsName tests fail.
func (suite *APIImageTagSuite) TestImageTagFailToUseSha256AsName(c *check.C) {
	repo, tag := "localhost:5000/sha256", "1.25"

	resp, err := suite.doTag(busyboxImage, repo, tag)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 400)
}

func (suite *APIImageTagSuite) doTag(source, repo, tag string) (*http.Response, error) {
	q := url.Values{}
	q.Set("repo", repo)
	q.Set("tag", tag)

	return request.Post(fmt.Sprintf("/images/%s/tag", source), request.WithQuery(q))
}

func (suite *APIImageTagSuite) checkTagReferenceExisting(c *check.C, tagRef string, ok bool) {
	resp, err := request.Get(fmt.Sprintf("/images/%s/json", tagRef))
	c.Assert(err, check.IsNil)

	status := http.StatusOK
	if !ok {
		status = http.StatusNotFound
	}
	CheckRespStatus(c, resp, status)
}
