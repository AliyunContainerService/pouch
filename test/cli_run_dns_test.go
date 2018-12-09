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
func (suite *PouchRunDNSSuite) TestRunWithUserDefinedNetwork(c *check.C) {
	cname := "TestRunWithUserDefinedNetwork"

	// Create a user-defined network
	command.PouchRun("network", "create", "--name", cname,
		"-d", "bridge",
		"--gateway", testGateWay,
		"--subnet", testSubnet).Assert(c, icmd.Success)
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
func (suite *PouchRunDNSSuite) TestRunWithBridgeNetwork(c *check.C) {
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

// TestRunWithDNSFlags tests DNS related flags.
func (suite *PouchRunDNSSuite) TestRunWithDNSFlags(c *check.C) {
	cname := "TestRunWithDNSFlags"

	res := command.PouchRun("run", "--name", cname,
		"--dns", "1.2.3.4",
		"--dns-option", "timeout:3",
		"--dns-search", "example.com",
		busyboxImage,
		"cat", "/etc/resolv.conf")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is correct in container
	out := strings.Trim(res.Stdout(), "\n")
	c.Assert(strings.Contains(out, "nameserver 1.2.3.4"), check.Equals, true)
	c.Assert(strings.Contains(out, "options timeout:3"), check.Equals, true)
	c.Assert(strings.Contains(out, "search example.com"), check.Equals, true)

	// test if the value is in inspect result
	dns, err := inspectFilter(cname, ".HostConfig.DNS")
	c.Assert(err, check.IsNil)
	c.Assert(dns, check.Equals, "[1.2.3.4]")

	dnsOptions, err := inspectFilter(cname, ".HostConfig.DNSOptions")
	c.Assert(err, check.IsNil)
	c.Assert(dnsOptions, check.Equals, "[timeout:3]")

	dnsSearch, err := inspectFilter(cname, ".HostConfig.DNSSearch")
	c.Assert(err, check.IsNil)
	c.Assert(dnsSearch, check.Equals, "[example.com]")
}

// TestRunWithDNSRepeatFlags tests repeated DNS related flags.
func (suite *PouchRunDNSSuite) TestRunWithDNSRepeatFlags(c *check.C) {
	cname := "TestRunWithDNSRepeatFlags"

	res := command.PouchRun("run", "--name", cname,
		"--dns", "1.2.3.4",
		"--dns", "2.3.4.5",
		"--dns-option", "timeout:3",
		"--dns-option", "ndots:9",
		"--dns-search", "mydomain",
		"--dns-search", "mydomain2",
		busyboxImage,
		"cat", "/etc/resolv.conf")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is correct in container
	out := strings.Trim(res.Stdout(), "\n")
	c.Assert(strings.Contains(out, "nameserver 1.2.3.4\nnameserver 2.3.4.5"), check.Equals, true)
	c.Assert(strings.Contains(out, "options timeout:3 ndots:9"), check.Equals, true)
	c.Assert(strings.Contains(out, "search mydomain mydomain2"), check.Equals, true)

	// test if the value is in inspect result
	dns, err := inspectFilter(cname, ".HostConfig.DNS")
	c.Assert(err, check.IsNil)
	c.Assert(dns, check.Equals, "[1.2.3.4 2.3.4.5]")

	dnsOptions, err := inspectFilter(cname, ".HostConfig.DNSOptions")
	c.Assert(err, check.IsNil)
	c.Assert(dnsOptions, check.Equals, "[timeout:3 ndots:9]")

	dnsSearch, err := inspectFilter(cname, ".HostConfig.DNSSearch")
	c.Assert(err, check.IsNil)
	c.Assert(dnsSearch, check.Equals, "[mydomain mydomain2]")
}
