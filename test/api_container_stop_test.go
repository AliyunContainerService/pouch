package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerStopSuite is the test suite for container stop API.
type APIContainerStopSuite struct{}

func init() {
	check.Suite(&APIContainerStopSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerStopSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestStopOk tests a running container could be stopped.
func (suite *APIContainerStopSuite) TestStopOk(c *check.C) {
	// must required
	cname := "TestStopOk"
	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"Cmd":        [1]string{"top"},
		"HostConfig": map[string]interface{}{},
	}

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	resp, err = request.Get(c, "/containers/"+cname+"/json")
	c.Assert(resp.StatusCode, check.Equals, 200, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/start")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/stop")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Delete(c, "/containers/"+cname)
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}

// TestNonExistingContainer tests stop a non-existing container return 404.
func (suite *APIContainerStopSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Post(c, "/containers/"+cname+"/stop")
	c.Assert(resp.StatusCode, check.Equals, 404, err.Error())
}

// TestInvalidParam tests using invalid parameter return.
func (suite *APIContainerStopSuite) TestInvalidParam(c *check.C) {
	//TODO
}
