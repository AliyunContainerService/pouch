package main

import (
	"encoding/json"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// PouchAPIContainerInspectSuite is the test suite for container inspect API.
type PouchAPIContainerInspectSuite struct{}

func init() {
	check.Suite(&PouchAPIContainerInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIContainerInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestInspectNoSuchContainer tests inspecting a container that doesn't exits return error.
func (suite *PouchAPIContainerInspectSuite) TestInspectNoSuchContainer(c *check.C) {
	resp, err := request.Get("/containers/nosuchcontainerxxx/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)
}

// TestInspectOk tests inspecting an existing container is OK.
func (suite *PouchAPIContainerInspectSuite) TestInpectOk(c *check.C) {
	// must required
	cname := "api-inspect-test"
	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	resp, err := request.Post("/containers/create", request.WithQuery(q),
		request.WithJSONBody(obj), request.WithHeader("Content-Type", "application/json"))
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	got := types.ContainerJSON{}
	err = json.NewDecoder(resp.Body).Decode(&got)
	c.Assert(err, check.IsNil)

	c.Assert(got.Image, check.Equals, busyboxImage)
	c.Assert(got.Name, check.Equals, cname)
}
