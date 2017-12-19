package main

import (
	"encoding/json"
	"net/http"
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

	// must required
	cname := "api-create-test"
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

	// Decode response
	got := types.ContainerCreateResp{}
	err = json.NewDecoder(resp.Body).Decode(&got)
	c.Assert(err, check.IsNil)

	c.Assert(got.ID, check.NotNil)

	resp, err = request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}

// TestCreateWithEmptyBody tests /containers/create API.
// If create request has an empty body, daemon should return HTTP status code of bad request.
func (suite *APIContainerCreateSuite) TestCreateWithEmptyBody(c *check.C) {
	config := map[string]interface{}{}
	resp, err := request.Post("/containers/create", request.WithJSONBody(config))

	c.Assert(err, check.IsNil)
	defer resp.Body.Close()

	c.Assert(resp.StatusCode, check.Equals, http.StatusBadRequest)
	// check the error message
	// HostConfig in request body cannot be nil

}
