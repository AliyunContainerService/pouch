package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerRenameSuite is the test suite for container create API.
type APIContainerRenameSuite struct{}

func init() {
	check.Suite(&APIContainerRenameSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerRenameSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestRenameOk test create api is ok with default parameters.
func (suite *APIContainerRenameSuite) TestRenameOk(c *check.C) {

	// must required
	oldname := "TestRenameOk"
	newname := "NewTestRenameOk"

	q := url.Values{}
	q.Add("name", oldname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)

	resp, err := request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	resp, err = request.Get(c, "/containers/"+oldname+"/json")
	c.Assert(resp.StatusCode, check.Equals, 200, err.Error())

	newq := url.Values{}
	newq.Add("name", newname)
	query = request.WithQuery(newq)
	resp, err = request.Post(c, "/containers/"+oldname+"/rename", query)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Delete(c, "/containers/"+newname)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}
