package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// PouchAPIContainerStopSuite is the test suite for container stop API.
type PouchAPIContainerStopSuite struct{}

func init() {
	check.Suite(&PouchAPIContainerStopSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIContainerStopSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestStopOk tests a running container could be stopped.
func (suite *PouchAPIContainerStopSuite) TestStopOk(c *check.C) {
	// must required
	cname := "TestStopOk"
	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"Cmd":        [1]string{"top"},
		"HostConfig": map[string]interface{}{},
	}

	resp, err := request.Post("/containers/create", request.WithQuery(q),
		request.WithJSONBody(obj), request.WithHeader("Content-Type", "application/json"))
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	resp, err = request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Post("/containers/" + cname + "/stop")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}
