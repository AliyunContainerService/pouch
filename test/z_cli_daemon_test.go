package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/daemon"
	"github.com/alibaba/pouch/test/environment"

	"github.com/alibaba/pouch/test/util"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchDaemonSuite is the test suite fo daemon.
type PouchDaemonSuite struct{}

func init() {
	check.Suite(&PouchDaemonSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchDaemonSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestDaemonCgroupParent tests daemon with cgroup parent
func (suite *PouchDaemonSuite) TestDaemonCgroupParent(c *check.C) {
	dcfg, err := StartDefaultDaemonDebug("--cgroup-parent=tmp")
	if err != nil {
		c.Skip("deamon start failed")
	}

	// Must kill it, as we may loose the pid in next call.
	defer dcfg.KillDaemon()

	cname := "TestDaemonCgroupParent"
	{
		result := command.PouchRun("--host", daemon.Listen, "pull", busyboxImage)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("pull image failed, err:%v", result)
		}
	}
	{
		result := command.PouchRun("--host", daemon.Listen, "run", "--name", cname, busyboxImage)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("run container failed, err:%v", result)
		}
	}
	defer DelContainerForceMultyTime(c, cname)

	// test if the value is in inspect result
	output := command.PouchRun("inspect", "--host", daemon.Listen, cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	//// test if cgroup has the right parent path
	//path := fmt.Sprintf("/sys/fs/cgroup/memory/tmp/%s", result.ID)
	//_, err = os.Stat(path)
	//if err != nil {
	//	daemon.DConfig.DumpLog()
	//	c.Fatalf("get cgroup path failed, err:%s", err)
	//}
}

// TestDaemonListenTCP tests daemon listen with tcp address.
func (suite *PouchDaemonSuite) TestDaemonListenTCP(c *check.C) {
	// Start a test daemon with test args.
	listeningPorts := [][]string{
		{"0.0.0.0", "0.0.0.0", "1236"},
		{"127.0.0.1", "127.0.0.1", "1234"},
		{"localhost", "127.0.0.1", "1235"},
	}

	for _, hostDirective := range listeningPorts {
		addr := fmt.Sprintf("tcp://%s:%s", hostDirective[0], hostDirective[2])
		dcfg := daemon.NewConfig()
		dcfg.Listen = ""
		dcfg.NewArgs("--listen=" + addr)
		err := dcfg.StartDaemon()
		c.Assert(err, check.IsNil)

		// verify listen to tcp works
		result := command.PouchRun("--host", addr, "version")
		dcfg.KillDaemon()
		result.Assert(c, icmd.Success)
	}
}

// TestDaemonConfigFile tests start daemon with configure file works.
func (suite *PouchDaemonSuite) TestDaemonConfigFile(c *check.C) {
	configFile := "/tmp/pouch.json"
	file, err := os.Create(configFile)
	c.Assert(err, check.IsNil)
	defer file.Close()
	defer os.Remove(configFile)

	// Unmarshal config.Config, all fields in this struct could be handled in configuration file.
	cfg := config.Config{
		Debug: true,
	}
	s, _ := json.Marshal(cfg)
	fmt.Fprintf(file, "%s", s)
	file.Sync()

	// TODO: uncomment this when issue #1003 is fixed.
	//dcfg, err := StartDefaultDaemonDebug("--config-file="+configFile)
	//{
	//	err := dcfg.StartDaemon()
	//	c.Assert(err, check.IsNil)
	//}
	//
	//// TODO: verify more
	//
	//// Must kill it, as we may loose the pid in next call.
	//defer dcfg.KillDaemon()

	// config file cowork with parameter, no confilct
}

// TestDaemonConfigFileConfilct tests start daemon with configure file confilicts with parameter.
func (suite *PouchDaemonSuite) TestDaemonConfigFileConfilct(c *check.C) {
	path := "/tmp/pouch.json"
	cfg := struct {
		ContainerdPath string `json:"containerd-path"`
	}{
		ContainerdPath: "abc",
	}
	err := CreateConfigFile(path, cfg)
	c.Assert(err, check.IsNil)
	defer os.Remove(path)

	dcfg, err := StartDefaultDaemon("--containerd-path", "def", "--config-file="+path)
	dcfg.KillDaemon()
	c.Assert(err, check.NotNil)
}

// TestDaemonConfigFileUnknownFlag tests start daemon with unknown flags in configure file.
func (suite *PouchDaemonSuite) TestDaemonConfigFileUnknownFlag(c *check.C) {
	path := "/tmp/pouch.json"
	cfg := struct {
		Adsj string `json:"adsj"`
	}{
		Adsj: "xxx",
	}
	err := CreateConfigFile(path, cfg)
	c.Assert(err, check.IsNil)
	defer os.Remove(path)

	dcfg, err := StartDefaultDaemon("--debug", "--config-file="+path)
	c.Assert(err, check.NotNil)
	dcfg.KillDaemon()
}

// TestDaemonConfigFileAndCli tests start daemon with configure file and CLI .
func (suite *PouchDaemonSuite) TestDaemonConfigFileAndCli(c *check.C) {
	// Check default configure file could work

	// TODO: uncomment if issue #1003 is fixed
	//path := "/etc/pouch/config.json"
	//cfg := struct {
	//	Labels []string `json:"labels,omitempty"`
	//}{
	//	Labels: []string{"a=b"},
	//}
	//err := CreateConfigFile(path, cfg)
	//c.Assert(err, check.IsNil)
	//defer os.Remove(path)
	//
	//// Do Not specify configure file explicitly, it should work.
	//dcfg, err := StartDefaultDaemonDebug()
	//c.Assert(err, check.IsNil)
	//defer dcfg.KillDaemon()
	//
	//result := RunWithSpecifiedDaemon(dcfg, "info")
	//err = util.PartialEqual(result.Stdout(), "a=b")
	//c.Assert(err, check.IsNil)
}

// TestDaemonInvalideArgs tests invalid args in deamon return error
func (suite *PouchDaemonSuite) TestDaemonInvalideArgs(c *check.C) {
	_, err := StartDefaultDaemon("--config=xxx")
	c.Assert(err, check.NotNil)
}

// TestDaemonRestart tests daemon restart with running container.
func (suite *PouchDaemonSuite) TestDaemonRestart(c *check.C) {
	dcfg, err := StartDefaultDaemonDebug()
	// Start a test daemon with test args.
	if err != nil {
		c.Skip("deamon start failed.")
	}
	// Must kill it, as we may loose the pid in next call.
	defer dcfg.KillDaemon()

	{
		result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("pull image failed, err:%v", result)
		}
	}

	cname := "TestDaemonRestart"
	{
		result := RunWithSpecifiedDaemon(dcfg, "run", "-d", "--name", cname,
			"-p", "1234:80",
			busyboxImage)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("run container failed, err:%v", result)
		}
	}
	defer DelContainerForceMultyTime(c, cname)

	// restart daemon
	err = RestartDaemon(dcfg)
	c.Assert(err, check.IsNil)

	// test if the container is running.
	output := RunWithSpecifiedDaemon(dcfg, "inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Fatalf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(result[0].State.Status), check.Equals, "running")
}

// TestDaemonLabel tests start daemon with label works.
func (suite *PouchDaemonSuite) TestDaemonLabel(c *check.C) {
	dcfg, err := StartDefaultDaemonDebug("--label", "a=b")
	// Start a test daemon with test args.
	if err != nil {
		c.Skip("deamon start failed.")
	}
	// Must kill it, as we may loose the pid in next call.
	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "info")
	err = util.PartialEqual(result.Stdout(), "a=b")
	c.Assert(err, check.IsNil)
}

// TestDaemonLabelDup tests start daemon with duplicated label works.
func (suite *PouchDaemonSuite) TestDaemonLabelDup(c *check.C) {
	dcfg, err := StartDefaultDaemonDebug("--label", "a=b", "--label", "a=b")
	// Start a test daemon with test args.
	if err != nil {
		c.Skip("deamon start failed.")
	}
	// Must kill it, as we may loose the pid in next call.
	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "info")
	err = util.PartialEqual(result.Stdout(), "a=b")
	c.Assert(err, check.IsNil)

	cnt := strings.Count(result.Stdout(), "a=b")
	c.Assert(cnt, check.Equals, 1)
}

// TestDaemonLabelNeg tests start daemon with wrong label could not work.
func (suite *PouchDaemonSuite) TestDaemonLabelNeg(c *check.C) {
	_, err := StartDefaultDaemon("--label", "adsf")
	c.Assert(err, check.NotNil)
}

// TestDaemonDefaultRegistry tests set default registry works.
func (suite *PouchDaemonSuite) TestDaemonDefaultRegistry(c *check.C) {
	dcfg, err := StartDefaultDaemonDebug(
		"--default-registry",
		"reg.docker.alibaba-inc.com",
		"--default-registry-namespace",
		"base")
	c.Assert(err, check.IsNil)

	// Check pull image with default registry using the registry specified in daemon.
	result := RunWithSpecifiedDaemon(dcfg, "pull", "hello-world")
	err = util.PartialEqual(result.Combined(), "reg.docker.alibaba-inc.com/base/hello-world")
	c.Assert(err, check.IsNil)

	defer dcfg.KillDaemon()
}
