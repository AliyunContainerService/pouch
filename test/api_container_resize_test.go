package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerResizeSuite is the test suite for container upgrade API.
type APIContainerResizeSuite struct{}

func init() {
	check.Suite(&APIContainerResizeSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerResizeSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestContainerResizeOk is to verify resize container.
func (suite *APIContainerResizeSuite) TestContainerResizeOk(c *check.C) {
	cname := "TestContainerResizeOk"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	q := url.Values{}
	q.Add("h", "10")
	q.Add("w", "10")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/resize", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
}

// TestContainerResizeWithInvalidSize is to verify resize container with invalid size.
func (suite *APIContainerResizeSuite) TestContainerResizeWithInvalidSize(c *check.C) {
	cname := "TestContainerResizeWithInvalidSize"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	q := url.Values{}
	q.Add("h", "hi")
	q.Add("w", "wo")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/resize", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 400)
}

// TestResizeCreatedContainer is to verify resize a created container.
func (suite *APIContainerResizeSuite) TestResizeCreatedContainer(c *check.C) {
	cname := "TestResizeCreatedContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	q := url.Values{}
	q.Add("h", "10")
	q.Add("w", "10")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/resize", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}

// TestResizeStoppedContainer is to verify resize a stoped container.
func (suite *APIContainerResizeSuite) TestResizeStoppedContainer(c *check.C) {
	cname := "TestResizeStoppedContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)
	StopContainerOk(c, cname)

	q := url.Values{}
	q.Add("h", "10")
	q.Add("w", "10")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/resize", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}
