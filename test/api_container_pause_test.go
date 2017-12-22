package main

import (
	"net/url"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerPauseSuite is the test suite for container pause/unpause API.
type APIContainerPauseSuite struct{}

func init() {
	check.Suite(&APIContainerPauseSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerPauseSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestPauseUnpauseOk tests a running container could be paused and unpaused.
func (suite *APIContainerPauseSuite) TestPauseUnpauseOk(c *check.C) {
	// must required
	cname := "TestPauseUnpauseOk"
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

	resp, err = request.Post(c, "/containers/"+cname+"/pause")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	// TODO: Add state check
	//resp = request.Get(c, "/containers/" + cname + "/json")
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got := types.ContainerJSON{}
	//err = request.DecodeBody(c, resp.Body, &got)
	//c.Assert(got.State.Status, check.Equals, "paused")

	resp, err = request.Post(c, "/containers/"+cname+"/unpause")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	//resp = request.Get(c,"/containers/" + cname + "/json")
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got = types.ContainerJSON{}
	//err = request.DecodeBody(c, resp.Body, &got)
	//c.Assert(got.State.Status, check.Equals, "running")

	// Need to set force=true in url.rawquery, as container's state is running
	q = url.Values{}
	q.Add("force", "true")
	resp, err = request.Delete(c, "/containers/"+cname, request.WithQuery(q))
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}

// TestNonExistingContainer tests pause a non-existing container return 404.
func (suite *APIContainerPauseSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Post(c, "/containers/"+cname+"/pause")
	c.Assert(resp.StatusCode, check.Equals, 404, err.Error())
}

// TestNotRunningContainer tests pausing a non-running container will return error.
func (suite *APIContainerPauseSuite) TestNotRunningContainer(c *check.C) {
	// must required
	cname := "TestNotRunningContainer"
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

	resp, err = request.Post(c, "/containers/"+cname+"/pause")
	c.Assert(resp.StatusCode, check.Equals, 500, err.Error())

	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got := types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "created")

	resp, err = request.Post(c, "/containers/"+cname+"/start")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/pause")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())

	resp, err = request.Post(c, "/containers/"+cname+"/pause")
	c.Assert(resp.StatusCode, check.Equals, 500, err.Error())

	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got = types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "paused")

	resp, err = request.Post(c, "/containers/"+cname+"/unpause")
	c.Assert(resp.StatusCode, check.Equals, 204)

	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got = types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "running")

	resp, err = request.Post(c, "/containers/"+cname+"/stop")
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Post(c, "/containers/"+cname+"/pause")
	c.Assert(resp.StatusCode, check.Equals, 500)

	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got = types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "exited")

	// Need to set force=true in url.rawquery, as container's state is running
	q = url.Values{}
	q.Add("force", "true")
	query = request.WithQuery(q)
	resp, err = request.Delete(c, "/containers/"+cname, query)
	c.Assert(resp.StatusCode, check.Equals, 204)
}
