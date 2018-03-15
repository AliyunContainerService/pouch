package main

import (
	"encoding/json"
	"runtime"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchNetworkSuite is the test suite for network CLI.
type PouchNetworkSuite struct{}

func init() {
	check.Suite(&PouchNetworkSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchNetworkSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)

	// Remove all Containers, in case there are legacy containers connecting network.
	environment.PruneAllContainers(apiClient)
}

// TestNetworkInspectFormat tests the inspect format of network works.
func (suite *PouchNetworkSuite) TestNetworkInspectFormat(c *check.C) {
	output := command.PouchRun("network", "inspect", "bridge").Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	// inspect network name
	output = command.PouchRun("network", "inspect", "-f", "{{.Name}}", "bridge").Stdout()
	c.Assert(output, check.Equals, "bridge\n")
}

// TestNetworkDefault tests the creation of default bridge/none/host network.
func (suite *PouchNetworkSuite) TestNetworkDefault(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	// After pouchd is launched, default netowrk bridge is created
	// check the existence of default network: bridge
	command.PouchRun("network", "inspect", "bridge").Assert(c, icmd.Success)

	command.PouchRun("network", "inspect", "none").Assert(c, icmd.Success)

	command.PouchRun("network", "inspect", "host").Assert(c, icmd.Success)

	// Check the existence of link: p0
	icmd.RunCommand("ip", "link", "show", "dev", "p0").Assert(c, icmd.Success)

	{
		// Assign the none network to a container.
		expct := icmd.Expected{
			ExitCode: 0,
			Out:      "",
		}
		err := command.PouchRun("run", "--name", funcname, "--net", "none", busyboxImage, "ip", "r").Compare(expct)
		c.Assert(err, check.IsNil)

		DelContainerForceMultyTime(c, funcname)
	}
	{
		cmd := "ip r |grep default"
		routeOnHost := icmd.RunCommand("bash", "-c", cmd).Stdout()
		// Assign the host network to a container.
		expct := icmd.Expected{
			ExitCode: 0,
			Out:      routeOnHost,
		}
		err := command.PouchRun("run", "--name", funcname, "--net", "host", busyboxImage, "ip", "r").Compare(expct)
		c.Assert(err, check.IsNil)

		DelContainerForceMultyTime(c, funcname)
	}
}

// TestNetworkBridgeWorks tests bridge network works.
func (suite *PouchNetworkSuite) TestNetworkBridgeWorks(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	// Remove network in case there is legacy network which may impacts test.
	defer command.PouchRun("network", "remove", funcname)

	gateway := "192.168.4.1"
	subnet := "192.168.4.0/24"

	command.PouchRun("network", "create", "--name", funcname,
		"-d", "bridge",
		"--gateway", gateway,
		"--subnet", subnet).Assert(c, icmd.Success)
	command.PouchRun("network", "inspect", funcname).Assert(c, icmd.Success)

	{
		// Assign network to a container works
		expct := icmd.Expected{
			ExitCode: 0,
			Out:      "eth0",
		}
		err := command.PouchRun("run", "--name", funcname, "--net", funcname, busyboxImage, "ip", "link", "ls", "eth0").Compare(expct)
		c.Assert(err, check.IsNil)

		DelContainerForceMultyTime(c, funcname)
	}

	{
		// remove network should fail
		expct := icmd.Expected{
			ExitCode: 1,
			Err:      "has active endpoints",
		}
		command.PouchRun("run", "--name", funcname, "--net", funcname, busyboxImage, "top").Assert(c, icmd.Success)

		err := command.PouchRun("network", "remove", funcname).Compare(expct)
		c.Assert(err, check.IsNil)

	}
	{
		// remove container, then the veth device should also been removed
		DelContainerForceMultyTime(c, funcname)

		// get the ID of bridge to construct the bridge name.
		cmd := "pouch network list |grep " + funcname + "|awk '{print $1}'"
		id := icmd.RunCommand("bash", "-c", cmd).Stdout()

		// there should be no veth interface on this bridge
		cmd = "brctl show |grep br-" + id + " |grep veth"
		expct := icmd.Expected{
			ExitCode: 1,
		}
		err := icmd.RunCommand("bash", "-c", cmd).Compare(expct)
		c.Assert(err, check.IsNil)
	}
	{
		// container process exist, then the veth device should also been removed
		command.PouchRun("run", "--name", funcname, "--net", funcname, busyboxImage, "echo", "test").Assert(c, icmd.Success)

		// get the ID of bridge to construct the bridge name.
		cmd := "pouch network list |grep " + funcname + "|awk '{print $1}'"
		id := icmd.RunCommand("bash", "-c", cmd).Stdout()

		// there should be no veth interface on this bridge
		cmd = "brctl show |grep br-" + id + " |grep veth"
		expct := icmd.Expected{
			ExitCode: 1,
		}
		err := icmd.RunCommand("bash", "-c", cmd).Compare(expct)
		c.Assert(err, check.IsNil)

		DelContainerForceMultyTime(c, funcname)
	}
	{
		// running container is stopped, then the veth device should also been removed
		command.PouchRun("run", "--name", funcname, "--net", funcname, busyboxImage, "top").Assert(c, icmd.Success)
		command.PouchRun("stop", funcname).Assert(c, icmd.Success)

		// get the ID of bridge to construct the bridge name.
		cmd := "pouch network list |grep " + funcname + "|awk '{print $1}'"
		id := icmd.RunCommand("bash", "-c", cmd).Stdout()

		// there should be no veth interface on this bridge
		cmd = "brctl show |grep br-" + id + "|grep veth"
		expct := icmd.Expected{
			ExitCode: 1,
		}
		err := icmd.RunCommand("bash", "-c", cmd).Compare(expct)
		c.Assert(err, check.IsNil)

		DelContainerForceMultyTime(c, funcname)
	}
	{
		// get the ID of bridge to construct the bridge name.
		cmd := "pouch network list |grep " + funcname + "|awk '{print $1}'"
		id := icmd.RunCommand("bash", "-c", cmd).Stdout()

		// remove network, brctl show should not have this bridge
		command.PouchRun("network", "remove", funcname).Assert(c, icmd.Success)
		cmd = "brctl show |grep br-" + id
		expct := icmd.Expected{
			ExitCode: 1,
		}
		err := icmd.RunCommand("bash", "-c", cmd).Compare(expct)
		c.Assert(err, check.IsNil)
	}
}

// TestNetworkCreateWrongDriver tests using wrong driver returns error.
func (suite *PouchNetworkSuite) TestNetworkCreateWrongDriver(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	expct := icmd.Expected{
		ExitCode: 1,
		Err:      "not found",
	}

	err := command.PouchRun("network", "create", "--name", funcname, "--driver", "wrongdriver").Compare(expct)
	c.Assert(err, check.IsNil)

	command.PouchRun("network", "remove", funcname)
}

// TestNetworkCreateWithLabel tests creating network with label.
func (suite *PouchNetworkSuite) TestNetworkCreateWithLabel(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	gateway := "192.168.3.1"
	subnet := "192.168.3.0/24"

	command.PouchRun("network", "create",
		"--name", funcname,
		"-d", "bridge",
		"--gateway", gateway,
		"--subnet", subnet,
		"--label", "test=foo").Assert(c, icmd.Success)
	command.PouchRun("network", "remove", funcname)
}

// TestNetworkCreateWithOption tests creating network with option.
func (suite *PouchNetworkSuite) TestNetworkCreateWithOption(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	gateway := "192.168.100.1"
	subnet := "192.168.100.0/24"

	command.PouchRun("network", "create",
		"--name", funcname,
		"-d", "bridge",
		"--gateway", gateway,
		"--subnet", subnet,
		"--option", "test=foo").Assert(c, icmd.Success)
	command.PouchRun("network", "remove", funcname)
}

// TestNetworkCreateDup tests creating duplicate network return error.
func (suite *PouchNetworkSuite) TestNetworkCreateDup(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	expct := icmd.Expected{
		ExitCode: 1,
		Err:      "already exist",
	}

	gateway1 := "192.168.101.1"
	subnet1 := "192.168.101.0/24"
	gateway2 := "192.168.102.1"
	subnet2 := "192.168.102.0/24"

	command.PouchRun("network", "create",
		"--name", funcname,
		"-d", "bridge",
		"--gateway", gateway1,
		"--subnet", subnet1).Assert(c, icmd.Success)

	err := command.PouchRun("network", "create",
		"--name", funcname,
		"-d", "bridge",
		"--gateway", gateway2,
		"--subnet", subnet2).Compare(expct)
	c.Assert(err, check.IsNil)

	command.PouchRun("network", "remove", funcname)
}

func (suite *PouchNetworkSuite) TestNetworkPortMapping(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	ret := icmd.RunCommand("which", "curl")
	if ret.ExitCode != 0 {
		c.Skip("Host does not have curl")
	}

	expct := icmd.Expected{
		ExitCode: 0,
		Out:      "It works",
	}

	image := "registry.hub.docker.com/library/httpd"

	command.PouchRun("pull", image).Assert(c, icmd.Success)
	command.PouchRun("run", "-d",
		"--name", funcname,
		"-p", "9999:80",
		image).Assert(c, icmd.Success)

	time.Sleep(1 * time.Second)
	err := icmd.RunCommand("curl", "localhost:9999").Compare(expct)
	c.Assert(err, check.IsNil)

	command.PouchRun("rm", "-f", funcname)
}
