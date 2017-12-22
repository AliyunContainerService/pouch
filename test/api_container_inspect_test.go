package main

import (
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
	resp, err := request.Get(c, "/containers/nosuchcontainerxxx/json")
	c.Assert(resp.StatusCode, check.Equals, 404, err.Error())
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
	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(c, "/containers/create", query, body)

	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	resp, err = request.Get(c, "/containers/"+cname+"/json")
	c.Assert(resp.StatusCode, check.Equals, 200, err.Error())

	got := types.ContainerJSON{}
	request.DecodeToStruct(c, resp.Body, &got)

	c.Assert(got.Image, check.Equals, busyboxImage)
	c.Assert(got.Name, check.Equals, cname)

	resp, err = request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}

// TestNonExistingContainer tests inspect a non-existing container return 404.
func (suite *APIContainerInspectSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Get(c, "/containers/"+cname+"/json")
	c.Assert(resp.StatusCode, check.Equals, 404, err.Error())
}

// TestRespValid tests the response of inspect is accurate.
func (suite *APIContainerInspectSuite) TestRespValid(c *check.C) {
	// TODO
}
