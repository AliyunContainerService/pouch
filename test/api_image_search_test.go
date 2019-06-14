package main

import (
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
)

// APIImageSearchSuite is the test suite for image search API.
type APIImageSearchSuite struct{}

func init() {
	check.Suite(&APIImageSearchSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageSearchSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

func (suite *APIImageSearchSuite) TestImageSearchOK(c *check.C) {
	q := url.Values{}
	q.Add("term", "nginx")
	q.Add("registry", "")

	query := request.WithQuery(q)
	resp, err := request.Get("/images/search", query)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 200)
	c.Assert(err, check.IsNil)

	var got []types.SearchResultItem
	request.DecodeBody(&got, resp.Body)

	c.Assert(util.PartialEqual(got[0].Name, "nginx"), check.IsNil)
	c.Assert(got[0].Description, check.NotNil)
	c.Assert(got[0].IsOfficial, check.NotNil)
	c.Assert(got[0].IsAutomated, check.NotNil)
	c.Assert(got[0].StarCount, check.NotNil)
}
