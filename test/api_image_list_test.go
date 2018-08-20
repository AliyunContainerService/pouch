package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImageListSuite is the test suite for image list API.
type APIImageListSuite struct{}

func init() {
	check.Suite(&APIImageListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageListOk tests listing images is OK.
func (suite *APIImageListSuite) TestImageListOk(c *check.C) {
	resp, err := request.Get("/images/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
}

// TestImageListAll tests listing all images layers.
func (suite *APIImageListSuite) TestImageListAll(c *check.C) {
	q := url.Values{}
	q.Add("all", "true")
	query := request.WithQuery(q)

	resp, err := request.Get("/images/json", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	// TODO: missing case
	//
	//	check the fields
}

// TestImageListDigest tests listing images digest.
func (suite *APIImageListSuite) TestImageListDigest(c *check.C) {
	q := url.Values{}
	q.Add("digests", "true")
	query := request.WithQuery(q)

	resp, err := request.Get("/images/json", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
}

// TestImageListFilter tests listing images with filter.
func (suite *APIImageListSuite) TestImageListFilter(c *check.C) {
	// TODO: missing case
	helpwantedForMissingCase(c, "iamge api list filter cases")
}
