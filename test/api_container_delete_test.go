package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerDeleteSuite is the test suite for container delete API.
type APIContainerDeleteSuite struct{}

func init() {
	check.Suite(&APIContainerDeleteSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerDeleteSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestDeleteNonExisting tests deleting a non-existing container return error.
func (suite *APIContainerDeleteSuite) TestDeleteNonExisting(c *check.C) {
	cname := "TestDeleteNonExisting"
	resp, err := request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)
}

// TestDeleteRunningCon test deleting running container return 500.
func (suite *APIContainerDeleteSuite) TestDeleteRunningCon(c *check.C) {
	cname := "TestDeleteRunningCon"

	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	path := "/containers/create"
	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, query, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 500)

	q = url.Values{}
	q.Add("force", "true")
	query = request.WithQuery(q)

	resp, err = request.Delete("/containers/"+cname, query)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}

// TestDeletePausedCon test deleting paused container return 500.
func (suite *APIContainerDeleteSuite) TestDeletePausedCon(c *check.C) {
	cname := "TestDeleteRunningCon"

	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	path := "/containers/create"
	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, query, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 500)

	q = url.Values{}
	q.Add("force", "true")
	query = request.WithQuery(q)

	resp, err = request.Delete("/containers/"+cname, query)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}

// TestDeleteStoppedCon test deleting stopped container return 204.
func (suite *APIContainerDeleteSuite) TestDeleteStoppedCon(c *check.C) {
	cname := "TestDeleteRunningCon"

	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	path := "/containers/create"
	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, query, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

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

// TestDeleteCreatedCon test deleting created container return 204.
func (suite *APIContainerDeleteSuite) TestDeleteCreatedCon(c *check.C) {
	cname := "TestDeleteCreatedCon"

	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	path := "/containers/create"
	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, query, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Delete("/containers/" + cname)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}
