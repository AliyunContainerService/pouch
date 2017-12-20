package main

import (
	"encoding/json"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// PouchAPICreateSuite is the test suite for info related API.
type PouchAPICreateSuite struct{}

func init() {
	check.Suite(&PouchAPICreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPICreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestCreateOk test create api is ok with default parameters.
func (suite *PouchAPICreateSuite) TestCreateOk(c *check.C) {

	// must required
	q := url.Values{}
	q.Add("name", "api-create-test")

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
}
