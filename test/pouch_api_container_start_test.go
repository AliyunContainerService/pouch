package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// PouchAPIContainerStartSuite is the test suite for container pause/unpause API.
type PouchAPIContainerStartSuite struct{}

func init() {
	check.Suite(&PouchAPIContainerStartSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIContainerStartSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestStartOk tests a running container could be paused and unpaused.
func (suite *PouchAPIContainerStartSuite) TestStartOk(c *check.C) {
	// must required
	cname := "TestStartOk"
	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"Cmd":        [1]string{"top"},
		"HostConfig": map[string]interface{}{},
	}

	path := "/containers/create"
	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, query, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	path = "/containers/" + cname + "/json"
	resp, err = request.Get(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	path = "/containers/" + cname + "/start"
	resp, err = request.Post(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	path = "/containers/" + cname + "/json"
	resp, err = request.Get(path)
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
	path = "/containers/" + cname
	resp, err = request.Delete(path, query)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}
