package main

import (
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerCreateSuite is the test suite for container create API.
type APIContainerCreateSuite struct{}

func init() {
	check.Suite(&APIContainerCreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerCreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestCreateOk test create api is ok with default parameters.
func (suite *APIContainerCreateSuite) TestCreateOk(c *check.C) {
	cname := "TestCreateOk"

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

	// Decode response
	got := types.ContainerCreateResp{}
	request.DecodeToStruct(c, resp.Body, &got)

	resp, err = request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}

// TestNilName tests creating container without giving name should succeed.
func (suite *APIContainerCreateSuite) TestNilName(c *check.C) {
	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post(c, "/containers/create", body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	// Decode response
	got := types.ContainerCreateResp{}
	request.DecodeToStruct(c, resp.Body, &got)
	c.Assert(got.ID, check.NotNil)
	c.Assert(got.Name, check.NotNil)

	resp, err = request.Delete(c, "/containers/"+got.Name)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}

// TestDupContainer tests create a duplicate container, return 409.
func (suite *APIContainerCreateSuite) TestDupContainer(c *check.C) {
	cname := "TestDupContainer"
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

	// Create a duplicate container
	resp, err = request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 409, err.Error())

	resp, err = request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}

// TestBadParam tests using bad parameter return 400.
func (suite *APIContainerCreateSuite) TestBadParam(c *check.C) {
	// TODO
}

// TestNonExistingImg tests using non-existing image return 404.
func (suite *APIContainerCreateSuite) TestNonExistingImg(c *check.C) {
	cname := "TestNonExistingImg"
	q := url.Values{}
	q.Add("name", cname)
	obj := map[string]interface{}{
		"Image":      "non-existing",
		"HostConfig": map[string]interface{}{},
	}
	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)

	resp, err := request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 404, err.Error())
}
