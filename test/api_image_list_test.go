package main

import (
	"net/url"
	"reflect"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
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
	q := url.Values{}

	repoDigest := environment.BusyboxRepo + "@" + environment.BusyboxDigest
	repoTag := environment.BusyboxRepo + ":" + environment.BusyboxTag

	f := filters.NewArgs()
	f.Add("reference", repoTag)
	filterJSON, err := filters.ToParam(f)
	c.Assert(err, check.IsNil)

	q.Add("filters", filterJSON)
	query := request.WithQuery(q)
	resp, err := request.Get("/images/json", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := []types.ImageInfo{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(got, check.NotNil)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.NotNil)
	c.Assert(got[0].CreatedAt, check.NotNil)
	c.Assert(got[0].Config, check.NotNil)
	c.Assert(got[0].Architecture, check.NotNil)
	c.Assert(got[0].Size, check.NotNil)
	c.Assert(got[0].Os, check.NotNil)
	c.Assert(reflect.DeepEqual(got[0].RepoTags, []string{repoTag}), check.Equals, true)
	c.Assert(reflect.DeepEqual(got[0].RepoDigests, []string{repoDigest}), check.Equals, true)
}

// TestImageListInvalidFilter tests listing images with invalid filter.
func (suite *APIImageListSuite) TestImageListInvalidFilter(c *check.C) {
	q := url.Values{}
	f := filters.NewArgs()
	f.Add("after", busyboxImage)
	filterJSON, err := filters.ToParam(f)
	c.Assert(err, check.IsNil)
	q.Add("filters", filterJSON)
	query := request.WithQuery(q)
	resp, err := request.Get("/images/json", query)

	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}
