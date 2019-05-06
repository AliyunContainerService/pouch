package main

import (
	"encoding/json"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunNetworkSuite is the test suite for run CLI.
type PouchRunNetworkSuite struct{}

func init() {
	check.Suite(&PouchRunNetworkSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunNetworkSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TestRunWithPing is to verify run container with network ping public website.
func (suite *PouchRunNetworkSuite) TestRunWithPing(c *check.C) {
	name := "TestRunWithPing"

	res := command.PouchRun("run", "--name", name,
		busyboxImage, "ping", "-c", "3", "www.taobao.com")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)
}

// TestRunWithPublicAll is to verify run container with publish-all flag
func (suite *PouchRunNetworkSuite) TestRunWithPublishAll(c *check.C) {
	name := "TestRunWithPublishAll"

	command.PouchRun("run", "--name", name, "--expose", "8080", "-P", busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()
	containers := make([]types.ContainerJSON, 1)
	err := json.Unmarshal([]byte(output), &containers)
	if err != nil || len(containers) == 0 {
		c.Fatal("fail to format container json")
	}
	c.Assert(len(containers[0].NetworkSettings.Ports), check.Equals, 1)
	checkPortMapExists(c, containers[0].NetworkSettings.Ports, "8080/tcp")

	// multiple expose port case
	name1 := "TestRunMultipleWithPublishAll"
	command.PouchRun("run", "--name", name1, "--expose", "8081", "--expose", "8082", "-P", busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name1)

	output = command.PouchRun("inspect", name1).Stdout()
	containers = make([]types.ContainerJSON, 1)
	err = json.Unmarshal([]byte(output), &containers)
	if err != nil || len(containers) == 0 {
		c.Fatal("fail to format container json")
	}
	c.Assert(len(containers[0].NetworkSettings.Ports), check.Equals, 2)
	checkPortMapExists(c, containers[0].NetworkSettings.Ports, "8081/tcp")
	checkPortMapExists(c, containers[0].NetworkSettings.Ports, "8082/tcp")
}

func checkPortMapExists(c *check.C, portMap types.PortMap, port string) {
	portBs, ok := portMap[port]
	c.Assert(ok, check.Equals, true)
	c.Assert(len(portBs), check.Equals, 1)
}

// TestRunWithMacAddress is to verify run container with mac address
func (suite *PouchRunNetworkSuite) TestRunWithMacAddress(c *check.C) {
	cname := "TestRunWithMacAddress"
	macAddress := "02:42:c0:a8:05:10"

	command.PouchRun("run", "-d", "--name", cname, "--mac-address", macAddress, busyboxImage, "sleep", "1000").Assert(c, icmd.Success)
	defer command.PouchRun("rm", "-vf", cname)

	res := command.PouchRun("exec", cname, "ip", "addr", "show")
	res.Assert(c, icmd.Success)

	stdout := res.Stdout()
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, macAddress) {
			found = true
			break
		}
	}

	c.Assert(found, check.Equals, true)
}

// TestRunWithIP is to verify run container with ipv4 address
func (suite *PouchRunNetworkSuite) TestRunWithIP(c *check.C) {
	cname := "TestRunWithIP"
	// TODO: add ipv6 address
	ipv4 := "192.168.5.100"

	command.PouchRun("run", "-d", "--name", cname, "--ip", ipv4, busyboxImage, "sleep", "1000").Assert(c, icmd.Success)
	defer command.PouchRun("rm", "-vf", cname)

	res := command.PouchRun("exec", cname, "ip", "addr", "show")
	res.Assert(c, icmd.Success)

	stdout := res.Stdout()
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, ipv4) {
			found = true
			break
		}
	}

	c.Assert(found, check.Equals, true)
}
