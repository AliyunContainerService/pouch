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
	"github.com/vishvananda/netlink"
)

// PouchNetworkSuite is the test suite for network CLI.
type PouchNetworkSuite struct{}

func init() {
	check.Suite(&PouchNetworkSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchNetworkSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)

	// Remove all Containers, in case there are legacy containers connecting network.
	environment.PruneAllContainers(apiClient)
}

// TestNetworkInspectFormat tests the inspect format of network works.
func (suite *PouchNetworkSuite) TestNetworkInspectFormat(c *check.C) {
	output := command.PouchRun("network", "inspect", "bridge").Stdout()
	result := []types.NetworkInspectResp{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	// inspect network name
	output = command.PouchRun("network", "inspect", "-f", "{{.Name}}", "bridge").Stdout()
	c.Assert(output, check.Equals, "bridge\n")

	output = command.PouchRun("network", "inspect", "bridge").Stdout()
	network := []types.NetworkInspectResp{}
	if err := json.Unmarshal([]byte(output), &network); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	networkID := network[0].ID

	// inspect network name by ID
	output = command.PouchRun("network", "inspect", "-f", "{{.Name}}", networkID).Stdout()
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

	// After pouchd is launched, default network bridge is created
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

	defer DelContainerForceMultyTime(c, funcname)
	{
		// Assign network to a container works
		expct := icmd.Expected{
			ExitCode: 0,
			Out:      "eth0",
		}
		err := command.PouchRun("run", "--name", funcname, "--net", funcname, busyboxImage,
			"ip", "link", "ls", "eth0").Compare(expct)
		c.Assert(err, check.IsNil)

		DelContainerForceMultyTime(c, funcname)
	}

	{
		// remove network should fail
		expct := icmd.Expected{
			ExitCode: 1,
			Err:      "has active endpoints",
		}
		command.PouchRun("run", "-d", "--name", funcname, "--net", funcname, busyboxImage, "top").Assert(c, icmd.Success)

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
		command.PouchRun("run", "-d", "--name", funcname, "--net", funcname, busyboxImage, "top").Assert(c, icmd.Success)
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
	defer command.PouchRun("network", "remove", funcname)
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
	defer command.PouchRun("network", "remove", funcname)
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
	defer command.PouchRun("network", "remove", funcname)

	err := command.PouchRun("network", "create",
		"--name", funcname,
		"-d", "bridge",
		"--gateway", gateway2,
		"--subnet", subnet2).Compare(expct)
	c.Assert(err, check.IsNil)

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

	image := "registry.hub.docker.com/library/httpd:2"

	command.PouchRun("pull", image).Assert(c, icmd.Success)
	command.PouchRun("run", "-d",
		"--name", funcname,
		"-p", "9999:80",
		image).Assert(c, icmd.Success)

	time.Sleep(10 * time.Second)
	err := icmd.RunCommand("curl", "localhost:9999").Compare(expct)
	c.Assert(err, check.IsNil)

	command.PouchRun("rm", "-f", funcname)
}

func createBridge(bridgeName string) (netlink.Link, error) {
	br, err := netlink.LinkByName(bridgeName)
	if err == nil && br != nil {
		return br, nil
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	b := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(b); err != nil {
		return nil, err
	}

	br, err = netlink.LinkByName(bridgeName)
	if err != nil {
		return nil, err
	}

	return br, nil
}

// TestNetworkConnect is to verify the correctness of 'network connect' command.
func (suite *PouchNetworkSuite) TestNetworkConnect(c *check.C) {
	bridgeName := "p1"
	networkName := "net1"
	containerName := "connect-test"

	// create bridge device
	br, err := createBridge("p1")
	c.Assert(err, check.Equals, nil)
	defer netlink.LinkDel(br)

	// create bridge network
	command.PouchRun("network", "create",
		"-d", "bridge",
		"--subnet=172.18.0.0/24", "--gateway=172.18.0.1",
		"-o", "com.docker.network.bridge.name="+bridgeName, networkName).Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("network", "rm", networkName).Assert(c, icmd.Success)
	}()

	// create container
	command.PouchRun("run", "-d", "--name", containerName, busyboxImage, "top").Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("rm", "-f", containerName).Assert(c, icmd.Success)
	}()

	// connect a network
	command.PouchRun("network", "connect", networkName, containerName).Assert(c, icmd.Success)

	// inspect container check result
	ret := command.PouchRun("inspect", containerName)
	ret.Assert(c, icmd.Success)

	out := ret.Stdout()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "net1") {
			found = true
			break
		}
	}

	c.Assert(found, check.Equals, true)
}

// TestNetworkDisconnect is to verify the correctness of 'network disconnect' command.
func (suite *PouchNetworkSuite) TestNetworkDisconnect(c *check.C) {
	name := "TestNetworkDisconnect"

	command.PouchRun("run", "-d", "--name", name, "--net", "bridge", busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	inspectInfo := command.PouchRun("inspect", name).Stdout()
	metaJSON := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(inspectInfo), &metaJSON); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	if _, ok := metaJSON[0].NetworkSettings.Networks["bridge"]; !ok {
		c.Errorf("container network mode should be 'bridge'")
	}

	command.PouchRun("network", "disconnect", "bridge", name).Assert(c, icmd.Success)
	inspectInfo = command.PouchRun("inspect", name).Stdout()
	metaJSON = []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(inspectInfo), &metaJSON); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	if len(metaJSON[0].NetworkSettings.Networks) != 0 {
		c.Errorf("container network config should be empty")
	}

	// Check restart container is ok after disconnect network
	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
	command.PouchRun("start", name).Assert(c, icmd.Success)
}

// TestNetworkConnectWithRestart is to verify the 'network connect'
// and 'network disconnect' after restart daemon.
func (suite *PouchNetworkSuite) TestNetworkConnectWithRestart(c *check.C) {
	// start the test pouch daemon
	dcfg, err := StartDefaultDaemonDebug()
	if err != nil {
		c.Skip("daemon start failed.")
	}
	defer dcfg.KillDaemon()

	// pull image
	RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage).Assert(c, icmd.Success)

	bridgeName := "p1"
	networkName := "net1"
	containerName := "TestNetworkConnectWithRestart"

	// create bridge device
	br, err := createBridge("p1")
	c.Assert(err, check.Equals, nil)
	defer netlink.LinkDel(br)

	// create bridge network
	RunWithSpecifiedDaemon(dcfg, "network", "create",
		"-d", "bridge",
		"--subnet=172.18.0.0/24", "--gateway=172.18.0.1",
		"-o", "com.docker.network.bridge.name="+bridgeName, networkName).Assert(c, icmd.Success)
	defer func() {
		RunWithSpecifiedDaemon(dcfg, "network", "rm", networkName).Assert(c, icmd.Success)
	}()

	// create container
	RunWithSpecifiedDaemon(dcfg, "run", "-d", "--name", containerName, busyboxImage, "top").Assert(c, icmd.Success)
	defer func() {
		RunWithSpecifiedDaemon(dcfg, "rm", "-f", containerName).Assert(c, icmd.Success)
	}()

	// restart daemon
	err = RestartDaemon(dcfg)
	c.Assert(err, check.IsNil)

	// connect a network
	RunWithSpecifiedDaemon(dcfg, "network", "connect", networkName, containerName).Assert(c, icmd.Success)

	// inspect container check result
	ret := RunWithSpecifiedDaemon(dcfg, "inspect", containerName).Assert(c, icmd.Success)

	out := ret.Stdout()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "net1") {
			found = true
			break
		}
	}
	c.Assert(found, check.Equals, true)

	// restart daemon
	err = RestartDaemon(dcfg)
	c.Assert(err, check.IsNil)

	// disconnect a network
	RunWithSpecifiedDaemon(dcfg, "network", "disconnect", networkName, containerName).Assert(c, icmd.Success)

	// inspect container check result
	ret = RunWithSpecifiedDaemon(dcfg, "inspect", containerName).Assert(c, icmd.Success)

	out = ret.Stdout()
	found = false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "net1") {
			found = true
			break
		}
	}

	c.Assert(found, check.Equals, false)
}
