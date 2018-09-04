package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerCommitSuite is the test suite for container commit API.
type APIContainerCommitSuite struct{}

func init() {
	check.Suite(&APIContainerCommitSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerCommitSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestCommitContainerSuccessful test commit a container successfully
func (suite *APIContainerCommitSuite) TestCommitContainerSuccessful(c *check.C) {
	cname := "TestCommitContainerSuccessful"

	defer DelContainerForceMultyTime(c, cname)
	CreateBusyboxContainerOk(c, cname)
	StartContainerOk(c, cname)

	q := url.Values{}
	q.Set("container", cname)
	q.Set("repo", "foo")
	q.Set("tag", "bar")

	query := request.WithQuery(q)
	resp, err := request.Post("/commit", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	DelImageForceOk(c, "foo:bar")
}

// TestCommitContainerNotExist test commit a non-exist container should fail
func (suite *APIContainerCommitSuite) TestCommitContainerNotExist(c *check.C) {
	cname := "TestCommitContainerNotExist"

	q := url.Values{}
	q.Set("container", cname)
	q.Set("repo", "fooo")
	q.Set("tag", "barr")

	query := request.WithQuery(q)
	resp, err := request.Post("/commit", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}
