package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APINetworkPruneSuite is the test suite for network prune API.
type APINetworkPruneSuite struct{}

func init() {
	check.Suite(&APINetworkPruneSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APINetworkPruneSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestNetworkPruneOk test if prune network is ok
func (suite *APINetworkPruneSuite) TestNetworkPruneOk(c *check.C) {
	network1 := "TestPruneNetwork1"
	command.PouchRun("network", "create", network1, "-d", "bridge", "--gateway", "192.168.1.1", "--subnet", "192.168.1.0/24").Assert(c, icmd.Success)

	network2 := "TestPruneNetwork2"
	command.PouchRun("network", "create", network2, "-d", "bridge", "--gateway", "192.168.2.1", "--subnet", "192.168.2.0/24").Assert(c, icmd.Success)

	// prune the network.
	path := "/networks/prune"
	resp, err := request.Post(path)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var networkPruneResp []string
	err = request.DecodeBody(&networkPruneResp, resp.Body)
	c.Assert(err, check.IsNil)
	c.Assert(len(networkPruneResp), check.Equals, 2)
}
