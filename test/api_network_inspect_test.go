package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APINetworkInspectSuite is the test suite for network inspect API.
type APINetworkInspectSuite struct{}

func init() {
	check.Suite(&APINetworkInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APINetworkInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestNetworkInspectOk tests if inspecting network is OK.
func (suite *APINetworkInspectSuite) TestNetworkInspectOk(c *check.C) {
	// Create a network.
	net := "TestNetworkInspectOk"
	obj := map[string]interface{}{
		"Name":   net,
		"Driver": "bridge",
		"IPAM": map[string]interface{}{
			"Config": []map[string]interface{}{
				{
					"Gateway": GateWay,
					"Subnet":  Subnet,
				},
			},
		},
	}

	path := "/networks/create"
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	// Inspect the created network.
	path = "/networks/" + net
	resp, err = request.Get(path)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	// Delete the network.
	path = "/networks/" + net
	resp, err = request.Delete(path)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)
}

// TestNetworkInspectNotExistent tests if inspecting non-existent network returns error.
func (suite *APINetworkInspectSuite) TestNetworkInspectNotExistent(c *check.C) {
	net := "NotExistentNetwork"

	path := "/networks/" + net
	resp, err := request.Get(path)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}
