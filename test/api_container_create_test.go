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
	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 201)

	// Decode response
	got := types.ContainerCreateResp{}
	request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)
	c.Assert(got.ID, check.NotNil)

	DelContainerForceOk(c, cname)
}

// TestNilName tests creating container without giving name should succeed.
func (suite *APIContainerCreateSuite) TestNilName(c *check.C) {
	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/create", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	// Decode response
	got := types.ContainerCreateResp{}
	request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)
	c.Assert(got.ID, check.NotNil)
	c.Assert(got.Name, check.NotNil)

	DelContainerForceOk(c, got.Name)
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

	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 201)

	// Create a duplicate container
	resp, err = request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 409)

	DelContainerForceOk(c, cname)
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

	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestBadParam tests using bad parameter return 400.
func (suite *APIContainerCreateSuite) TestBadParam(c *check.C) {

	// TODO:
	// 1. Invalid container name, for example length too large, illegal letter.
	// 2. Invalid Parameters
}

// TestAllocateTTY tests allocating tty is OK.
func (suite *APIContainerCreateSuite) TestAllocateTTY(c *check.C) {
	cname := "TestAllocateTTY"
	q := url.Values{}
	q.Add("name", cname)
	query := request.WithQuery(q)

	// HostConfig is nil.
	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"Tty":        true,
		"HostConfig": map[string]interface{}{},
	}

	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 201)

	// TODO: verify TTY works?
	DelContainerForceOk(c, cname)
}

// TestAddVolume tests add volume is OK.
func (suite *APIContainerCreateSuite) TestAddVolume(c *check.C) {
	cname := "TestAddVolume"
	q := url.Values{}
	q.Add("name", cname)
	query := request.WithQuery(q)

	// HostConfig is nil.
	obj := map[string]interface{}{
		"Image": busyboxImage,
		"Tty":   true,
		"HostConfig": map[string]interface{}{
			"Binds": [1]string{"/tmp:/tmp"},
		},
	}

	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 201)

	// TODO: verify volume
	DelContainerForceOk(c, cname)
}

// TestRuntime tests specify a different runtime, e.g. runv could work.
func (suite *APIContainerCreateSuite) TestRuntime(c *check.C) {
	cname := "TestRuntime"
	q := url.Values{}
	q.Add("name", cname)
	query := request.WithQuery(q)

	// HostConfig is nil.
	obj := map[string]interface{}{
		"Image": busyboxImage,
		"Tty":   true,
		"HostConfig": map[string]interface{}{
			"Runtime": "runv",
		},
	}

	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 201)

	// TODO: verify runtime
	DelContainerForceOk(c, cname)
}

// TestLxcfsEnable is OK.
func (suite *APIContainerCreateSuite) TestLxcfsEnable(c *check.C) {
	cname := "TestLxcfsEnable"
	q := url.Values{}
	q.Add("name", cname)
	query := request.WithQuery(q)

	isEnable := true

	obj := map[string]interface{}{
		"Image": busyboxImage,
		"HostConfig": map[string]interface{}{
			"EnableLxcfs": isEnable,
		},
	}

	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 201)

	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(got.HostConfig.EnableLxcfs, check.Equals, isEnable)

	DelContainerForceOk(c, cname)
}
