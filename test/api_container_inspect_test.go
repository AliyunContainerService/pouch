package main

import (
	"encoding/json"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerInspectSuite is the test suite for container inspect API.
type APIContainerInspectSuite struct{}

func init() {
	check.Suite(&APIContainerInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestInspectNoSuchContainer tests inspecting a container that doesn't exits return error.
func (suite *APIContainerInspectSuite) TestInspectNoSuchContainer(c *check.C) {
	resp, err := request.Get("/containers/nosuchcontainerxxx/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)
}

// TestInspectOk tests inspecting an existing container is OK.
func (suite *APIContainerInspectSuite) TestInpectOk(c *check.C) {
	// must required
	cname := "TestInpectOk"
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

	resp, err = request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}

// TestNonExistingContainer tests inspect a non-existing container return 404.
func (suite *APIContainerInspectSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)
}

// TestRespValid tests the response of inspect is accurate.
func (suite *APIContainerInspectSuite) TestRespValid(c *check.C) {
	// TODO
}
