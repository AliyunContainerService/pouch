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

	resp, err := request.Post("/containers/create", request.WithQuery(q),
		request.WithJSONBody(obj), request.WithHeader("Content-Type", "application/json"))
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	resp, err = request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	// TODO: Add state check
	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got := types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "paused")

	resp, err = request.Post("/containers/" + cname + "/unpause")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got = types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "running")

	// Need to set force=true in url.rawquery, as container's state is running
	q = url.Values{}
	q.Add("force", "true")
	resp, err = request.Delete("/containers/"+cname, request.WithQuery(q))
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}

// TestNonExistingContainer tests pause a non-existing container return 404.
func (suite *APIContainerPauseSuite) TestNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)
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

	resp, err := request.Post("/containers/create", request.WithQuery(q),
		request.WithJSONBody(obj), request.WithHeader("Content-Type", "application/json"))
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 500)

	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got := types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "created")

	resp, err = request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 500)

	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got = types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "paused")

	resp, err = request.Post("/containers/" + cname + "/unpause")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	//resp, err = request.Get("/containers/" + cname + "/json")
	//c.Assert(err, check.IsNil)
	//c.Assert(resp.StatusCode, check.Equals, 200)
	//got = types.ContainerJSON{}
	//err = request.DecodeBody(&got, resp.Body)
	//c.Assert(err, check.IsNil)
	//c.Assert(got.State.Status, check.Equals, "running")

	resp, err = request.Post("/containers/" + cname + "/stop")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)

	resp, err = request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)
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
	resp, err = request.Delete("/containers/"+cname, request.WithQuery(q))
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}
