package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerStartSuite is the test suite for container pause/unpause API.
type APIContainerStartSuite struct{}

func init() {
	check.Suite(&APIContainerStartSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerStartSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestStartOk tests starting container could work.
func (suite *APIContainerStartSuite) TestStartOk(c *check.C) {
	cname := "TestStartOk"
	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"Cmd":        [1]string{"top"},
		"HostConfig": map[string]interface{}{},
	}

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	resp, err = request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	// Comment until inspect interface returns this
	//got := &types.ContainerJSON{}
	//err = request.DecodeBody(got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Running, check.Equals, true)

	// Need to set force=true in url.rawquery, as container's state is running
	q = url.Values{}
	q.Add("force", "true")
	query = request.WithQuery(q)

	resp, err = request.Delete("/containers/"+cname, query)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}

// TestNonExistingContainer tests start a non-existing container return 404.
func (suite *APIContainerStartSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"

	resp, err := request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)
}

// TestInvalidParam tests using invalid parameter return.
func (suite *APIContainerStartSuite) TestInvalidParam(c *check.C) {
	//TODO
}
