package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APINetworkCreateSuite is the test suite for network create API.
type APINetworkCreateSuite struct{}

func init() {
	check.Suite(&APINetworkCreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APINetworkCreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestNetworkCreateOk tests creating network is OK.
func (suite *APINetworkCreateSuite) TestNetworkCreateOk(c *check.C) {
	nname := "TestNetworkCreateOk"
	obj := map[string]interface{}{
		"Name":   nname,
		"Driver": "bridge",
	}

	// TODO: issue #481 has been fixed, add IPAM config

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/networks/create", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	//DelNetworkOk(c, nname)
}

// TestNetworkCreateNilName tests creating network without name returns error.
func (suite *APINetworkCreateSuite) TestNetworkCreateNilName(c *check.C) {
	obj := map[string]interface{}{
		"Name": nil,
		"NetworkCreate": map[string]interface{}{
			"Driver": "bridge",
		},
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/networks/create", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}
