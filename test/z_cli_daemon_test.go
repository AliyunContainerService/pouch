package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/apis/types"
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
		result := command.PouchRun("--host", daemon.Listen, "run", "-d", "--name", cname, busyboxImage)
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

//// TestDaemonConfigFile tests start daemon with configure file works.
//func (suite *PouchDaemonSuite) TestDaemonConfigFile(c *check.C) {
//	path := "/tmp/pouch.json"
//
//	// Unmarshal config.Config, all fields in this struct could be handled in configuration file.
//	cfg := config.Config{
//		Debug: true,
//	}
//	err := CreateConfigFile(path, cfg)
//	c.Assert(err, check.IsNil)
//	defer os.Remove(path)
//
//	dcfg, err := StartDefaultDaemonDebug("--config-file=" + path)
//	{
//		err := dcfg.StartDaemon()
//		c.Assert(err, check.IsNil)
//	}
//
//	// Must kill it, as we may loose the pid in next call.
//	defer dcfg.KillDaemon()
//}

// TestDaemonConfigFileConflict tests start daemon with configure file conflicts with parameter.
func (suite *PouchDaemonSuite) TestDaemonConfigFileConflict(c *check.C) {
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

// TestDaemonNestObjectConflict tests start daemon with configure file contains nest objects conflicts with parameter.
func (suite *PouchDaemonSuite) TestDaemonNestObjectConflict(c *check.C) {
	path := "/tmp/pouch_nest.json"
	type TLSConfig struct {
		CA               string `json:"tlscacert,omitempty"`
		Cert             string `json:"tlscert,omitempty"`
		Key              string `json:"tlskey,omitempty"`
		VerifyRemote     bool   `json:"tlsverify"`
		ManagerWhiteList string `json:"manager-whitelist"`
	}
	cfg := struct {
		TLS TLSConfig
	}{
		TLS: TLSConfig{
			CA:   "ca",
			Cert: "cert",
			Key:  "key",
		},
	}
	err := CreateConfigFile(path, cfg)
	c.Assert(err, check.IsNil)
	defer os.Remove(path)

	dcfg, err := StartDefaultDaemon("--tlscacert", "ca", "--config-file="+path)
	dcfg.KillDaemon()
	c.Assert(err, check.NotNil)
}

// TestDaemonSliceFlagNotConflict tests start daemon with configure file contains slice flag will not conflicts with parameter.
func (suite *PouchDaemonSuite) TestDaemonSliceFlagNotConflict(c *check.C) {
	path := "/tmp/pouch_slice.json"
	cfg := struct {
		Labels []string `json:"label"`
	}{
		Labels: []string{"a=a", "b=b"},
	}
	err := CreateConfigFile(path, cfg)
	c.Assert(err, check.IsNil)
	defer os.Remove(path)

	dcfg, err := StartDefaultDaemon("--label", "c=d", "--config-file="+path)
	dcfg.KillDaemon()
	c.Assert(err, check.IsNil)
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
	path := "/etc/pouch/config.json"
	cfg := struct {
		Labels []string `json:"label,omitempty"`
	}{
		Labels: []string{"a=b"},
	}
	err := CreateConfigFile(path, cfg)
	c.Assert(err, check.IsNil)
	defer os.Remove(path)

	// Do Not specify configure file explicitly, it should work.
	dcfg, err := StartDefaultDaemonDebug()
	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "info")
	err = util.PartialEqual(result.Stdout(), "a=b")
	c.Assert(err, check.IsNil)
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
		c.Skip("daemon start failed.")
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
			busyboxImage, "top")
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

// TestDaemonRestartWithPausedContainer tests daemon with paused container.
func (suite *PouchDaemonSuite) TestDaemonRestartWithPausedContainer(c *check.C) {
	dcfg, err := StartDefaultDaemonDebug()
	//Start a test daemon with test args.
	if err != nil {
		c.Skip("daemon start failed")
	}
	defer dcfg.KillDaemon()

	{
		result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("pull image failed, err: %v", result)
		}
	}

	cname := "TestDaemonRestartWithPausedContainer"
	{
		result := RunWithSpecifiedDaemon(dcfg, "run", "-d", "--name", cname,
			"-p", "1234:80", busyboxImage, "top")
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("run container failed, err: %v", result)
		}

		// pause the container
		result = RunWithSpecifiedDaemon(dcfg, "pause", cname)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("pause container failed, err: %v", result)
		}
	}
	defer DelContainerForceMultyTime(c, cname)

	// restart daemon
	err = RestartDaemon(dcfg)
	c.Assert(err, check.IsNil)

	// test if the container is paused.
	output := RunWithSpecifiedDaemon(dcfg, "inspect", cname).Stdout()
	data := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		c.Fatalf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(data[0].State.Status), check.Equals, "paused")

	// unpause the container
	result := RunWithSpecifiedDaemon(dcfg, "unpause", cname)
	if result.ExitCode != 0 {
		dcfg.DumpLog()
		c.Fatalf("unpause container failed, err: %v", result)
	}

	//test if the container is running
	output = RunWithSpecifiedDaemon(dcfg, "inspect", cname).Stdout()
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		c.Fatalf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(data[0].State.Status), check.Equals, "running")
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

// TestDaemonTlsVerify tests start daemon with TLS verification enabled.
func (suite *PouchDaemonSuite) TestDaemonTlsVerify(c *check.C) {
	SkipIfFalse(c, IsTLSExist)
	dcfg := daemon.NewConfig()
	dcfg.Listen = ""
	dcfg.NewArgs("--listen=" + testDaemonHTTPSAddr)
	dcfg.Args = append(dcfg.Args,
		"--tlsverify",
		"--tlscacert="+serverCa,
		"--tlscert="+serverCert,
		"--tlskey="+serverKey)
	dcfg.Debug = false
	// Skip error check, because the function to check daemon up using CLI without TLS info.
	dcfg.StartDaemon()

	// Must kill it, as we may loose the pid in next call.
	defer dcfg.KillDaemon()

	// Use TLS could success
	result := RunWithSpecifiedDaemon(&dcfg,
		"--tlscacert="+clientCa,
		"--tlscert="+clientCert,
		"--tlskey="+clientKey, "version")
	result.Assert(c, icmd.Success)

	// Do not use TLS should fail
	result = RunWithSpecifiedDaemon(&dcfg, "version")
	c.Assert(result.ExitCode, check.Equals, 1)
	err := util.PartialEqual(result.Stderr(), "malformed HTTP response")
	c.Assert(err, check.IsNil)

	{
		// Use wrong CA should fail
		result := RunWithSpecifiedDaemon(&dcfg,
			"--tlscacert="+clientWrongCa,
			"--tlscert="+clientCert,
			"--tlskey="+clientKey, "version")
		c.Assert(result.ExitCode, check.Equals, 1)
		err := util.PartialEqual(result.Stderr(), "failed to append certificates")
		c.Assert(err, check.IsNil)
	}
}

// TestDaemonStartOverOneTimes tests start daemon over one times should fail.
func (suite *PouchDaemonSuite) TestDaemonStartOverOneTimes(c *check.C) {
	dcfg1 := daemon.NewConfig()
	dcfg1.Listen = ""
	addr1 := "unix:///var/run/pouchtest1.sock"
	dcfg1.NewArgs("--listen=" + addr1)
	err := dcfg1.StartDaemon()
	c.Assert(err, check.IsNil)

	// verify listen to tcp works
	command.PouchRun("--host", addr1, "version").Assert(c, icmd.Success)
	defer dcfg1.KillDaemon()

	// test second daemon with same pidfile should start fail
	dcfg2 := daemon.NewConfig()
	dcfg2.Listen = ""
	addr2 := "unix:///var/run/pouchtest2.sock"
	dcfg2.NewArgs("--listen=" + addr2)
	err = dcfg2.StartDaemon()
	c.Assert(err, check.NotNil)

}
