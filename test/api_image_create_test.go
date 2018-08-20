package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImageCreateSuite is the test suite for image create API.
type APIImageCreateSuite struct{}

func init() {
	check.Suite(&APIImageCreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageCreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageCreateOk tests creating an image is OK.
func (suite *APIImageCreateSuite) TestImageCreateOk(c *check.C) {
	q := url.Values{}
	q.Add("fromImage", environment.HelloworldRepo)
	q.Add("tag", "latest")

	query := request.WithQuery(q)
	resp, err := request.Post("/images/create", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
	c.Assert(discardPullStatus(resp.Body), check.IsNil)

	DelImageForceOk(c, environment.HelloworldRepo+":latest")
}

// TestImageCreateNil tests fromImage is nil.
func (suite *APIImageCreateSuite) TestImageCreateNil(c *check.C) {
	q := url.Values{}
	q.Add("fromImage", "")
	query := request.WithQuery(q)

	resp, err := request.Post("/images/create", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 400)
}

// TestImageCreateWithoutTag tests creating an image without tag, will use "latest" by default.
func (suite *APIImageCreateSuite) TestImageCreateWithoutTag(c *check.C) {
	q := url.Values{}
	q.Add("fromImage", environment.HelloworldRepo)
	query := request.WithQuery(q)

	resp, err := request.Post("/images/create", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
	c.Assert(discardPullStatus(resp.Body), check.IsNil)

	DelImageForceOk(c, environment.HelloworldRepo)
}

// TestImageCreateWithoutRegistry tests creating an image only by name, will use "latest" by default.
func (suite *APIImageCreateSuite) TestImageCreateWithoutRegistry(c *check.C) {
	q := url.Values{}
	q.Add("fromImage", helloworldImageOnlyRepoName)
	query := request.WithQuery(q)

	resp, err := request.Post("/images/create", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
	c.Assert(discardPullStatus(resp.Body), check.IsNil)

	DelImageForceOk(c, helloworldImageOnlyRepoName+":latest")
}
