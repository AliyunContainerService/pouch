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
	q.Add("fromImage", busyboxImage)
	query := request.WithQuery(q)
	resp, err := request.Post(c, "/images/create", query)
	c.Assert(resp.StatusCode, check.Equals, 200, err.Error())
}
