package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APINetworkListSuite is the test suite for network list API.
type APINetworkListSuite struct{}

func init() {
	check.Suite(&APINetworkListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APINetworkListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestNetworkListOk tests if inspecting network is OK.
func (suite *APINetworkListSuite) TestNetworkListOk(c *check.C) {
	// Create a network.
	net := "TestNetworkListOk"
	obj := map[string]interface{}{
		"Name":   net,
		"Driver": "bridge",
	}
	path := "/networks/create"
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	// List the created network.
	path = "/networks"
	resp, err = request.Get(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	// Delete the network.
	path = "/networks/" + net
	resp, err = request.Delete(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}
