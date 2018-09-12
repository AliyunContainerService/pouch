package main

import (
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerListSuite is the test suite for container list API.
type APIContainerListSuite struct{}

func init() {
	check.Suite(&APIContainerListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)
	PullImage(c, busyboxImage125)
}

// TestListOk test api is ok with default parameters.
func (suite *APIContainerListSuite) TestListOk(c *check.C) {
	cname := "TestListOk"

	resp, err := CreateBusybox125Container(cname, "top")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	defer DelContainerForceMultyTime(c, cname)

	// NOTE: need to list all here because the container is not running
	q := url.Values{}
	q.Set("all", "true")
	resp, err = request.Get("/containers/json", request.WithQuery(q))
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var got []types.Container
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(len(got), check.Equals, 1)

	// TODO: missing case
	//
	// add more field checker
	c.Assert(got[0].Names[0], check.Equals, cname)
	c.Assert(got[0].Image, check.Equals, busyboxImage125)
	c.Assert(got[0].ImageID, check.Equals, busyboxImage125ID)
}
