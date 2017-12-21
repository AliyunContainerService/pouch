package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// PouchAPIContainerRenameSuite is the test suite for container create API.
type PouchAPIContainerRenameSuite struct{}

func init() {
	check.Suite(&PouchAPIContainerRenameSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIContainerRenameSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestRenameOk test create api is ok with default parameters.
func (suite *PouchAPIContainerRenameSuite) TestRenameOk(c *check.C) {

	// must required
	oldname := "TestRenameOk"
	newname := "NewTestRenameOk"

	q := url.Values{}
	q.Add("name", oldname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	resp, err := request.Post("/containers/create", request.WithQuery(q),
		request.WithJSONBody(obj), request.WithHeader("Content-Type", "application/json"))
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Get("/containers/" + oldname + "/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	newq := url.Values{}
	newq.Add("name", newname)
	resp, err = request.Post("/containers/"+oldname+"/rename", request.WithQuery(newq))
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Delete("/containers/" + newname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}
