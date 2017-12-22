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
	resp, err := request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 404, err.Error())
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

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/start")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 500, err.Error())

	q = url.Values{}
	q.Add("force", "true")
	query = request.WithQuery(q)

	resp, err = request.Delete(c, "/containers/"+cname, query)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
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

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/start")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/pause")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 500, err.Error())

	q = url.Values{}
	q.Add("force", "true")
	query = request.WithQuery(q)

	resp, err = request.Delete(c, "/containers/"+cname, query)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
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

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/start")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/stop")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

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

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	resp, err = request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}
