package main

import (
	"net/url"
	"time"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/alibaba/pouch/test/util"
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

	// TODO: add a waituntil func to check the exsitence of image
	time.Sleep(5000 * time.Millisecond)

	resp, err = request.Delete("/images/" + environment.HelloworldRepo + ":latest")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
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

	time.Sleep(5000 * time.Millisecond)

	resp, err = request.Delete("/images/" + environment.HelloworldRepo + ":latest")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
}

// TestImageCreateWithoutRegistry tests creating an image only by name, will use "latest" by default.
func (suite *APIImageCreateSuite) TestImageCreateWithoutRegistry(c *check.C) {
	q := url.Values{}
	q.Add("fromImage", helloworldImageOnlyRepoName)
	query := request.WithQuery(q)
	resp, err := request.Post("/images/create", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	ret := util.WaitTimeout(10*time.Second, DelHelloworldImage)
	c.Assert(ret, check.Equals, true)
}

func DelHelloworldImage() bool {
	resp, err := request.Delete("/images/" + helloworldImageOnlyRepoName + ":latest")
	if err == nil && resp.StatusCode == 204 {
		return true
	}
	return false
}
