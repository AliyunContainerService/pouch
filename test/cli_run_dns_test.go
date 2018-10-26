package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunDNSSuite is the test suite for run CLI.
type PouchRunDNSSuite struct{}

func init() {
	check.Suite(&PouchRunDNSSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunDNSSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunDNSSuite) TearDownTest(c *check.C) {
}

// TestRunWithUserDefinedNetwork tests enabling libnetwork resolver if user-defined network.
func (suite *PouchRunSuite) TestRunWithUserDefinedNetwork(c *check.C) {
	cname := "TestRunWithUserDefinedNetwork"

	// Create a user-defined network
	command.PouchRun("network", "create", "--name", cname,
		"-d", "bridge",
		"--gateway", GateWay,
		"--subnet", Subnet).Assert(c, icmd.Success)
	defer command.PouchRun("network", "remove", cname)

	// Assign the user-defined network to a container
	res := command.PouchRun("run", "--name", cname,
		"--net", cname, busyboxImage,
		"cat", "/etc/resolv.conf")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	c.Assert(strings.Contains(res.Stdout(), "nameserver 127.0.0.11"), check.Equals, true)
}

// TestRunWithBridgeNetwork tests disabling libnetwork resolver if not user-defined network.
func (suite *PouchRunSuite) TestRunWithBridgeNetwork(c *check.C) {
	cname := "TestRunWithBridgeNetwork"

	// Use bridge network if not set --net.
	res := command.PouchRun("run", "--name", cname, busyboxImage,
		"cat", "/etc/resolv.conf")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	hostRes := icmd.RunCommand("cat", "/etc/resolv.conf")
	hostRes.Assert(c, icmd.Success)

	c.Assert(res.Stdout(), check.Equals, hostRes.Stdout())
}
