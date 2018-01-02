package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APINetworkDeleteSuite is the test suite for network delete API.
type APINetworkDeleteSuite struct{}

func init() {
	check.Suite(&APINetworkDeleteSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APINetworkDeleteSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestNetworkDelNonExisting tests deleting non-existing network.
func (suite *APINetworkDeleteSuite) TestNetworkDelNonExisting(c *check.C) {
	resp, err := request.Delete("/networks/" + "TestNetworkDelNonExisting")
	c.Assert(err, check.IsNil)
	// TODO: change to 404 once delete API has been implemented.
	CheckRespStatus(c, resp, 200)
}
