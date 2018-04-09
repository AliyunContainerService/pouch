package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/daemon"
	"github.com/alibaba/pouch/test/environment"

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
	// Start a test daemon with test args.
	daemon.DConfig = daemon.NewConfig()
	daemon.DConfig.Args = append(daemon.DConfig.Args, "--cgroup-parent=tmp")

	err := daemon.DConfig.StartDaemon()
	if err != nil {
		c.Skip("deamon start failed")
	}
	// Must kill it, as we may loose the pid in next call.
	defer daemon.DConfig.KillDaemon()

	cname := "TestDaemonCgroupParent"
	{
		result := command.PouchRun("--host", daemon.Listen, "pull", busyboxImage)
		if result.ExitCode != 0 {
			daemon.DConfig.DumpLog()
			c.Fatalf("pull image failed, err:%v", result)
		}
	}
	{
		result := command.PouchRun("--host", daemon.Listen, "run", "--name", cname, busyboxImage)
		if result.ExitCode != 0 {
			daemon.DConfig.DumpLog()
			c.Fatalf("run container failed, err:%v", result)
		}
	}
	defer DelContainerForceMultyTime(c, cname)

	// test if the value is in inspect result
	output := command.PouchRun("inspect", "--host", daemon.Listen, cname).Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
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
		{"0.0.0.0", "0.0.0.0", "5678"},
		{"127.0.0.1", "127.0.0.1", "1234"},
		{"localhost", "127.0.0.1", "1235"},
	}

	for _, hostDirective := range listeningPorts {
		daemon.DConfig = daemon.NewConfig()

		addr := fmt.Sprintf("tcp://%s:%s", hostDirective[0], hostDirective[2])
		daemon.DConfig.Listen = append(daemon.DConfig.Listen, addr)
		daemon.DConfig.Listen = append(daemon.DConfig.Listen, addr)
		daemon.DConfig.Args = append(daemon.DConfig.Args, "--listen="+addr)

		err := daemon.DConfig.StartDaemon()
		c.Assert(err, check.IsNil)

		// verify listen to tcp works
		command.PouchRun("--host", addr, "version").Assert(c, icmd.Success)

		daemon.DConfig.KillDaemon()
	}
}

// TestDaemonConfigFile tests start daemon with configfile works.
func (suite *PouchDaemonSuite) TestDaemonConfigFile(c *check.C) {
	configFile := "/tmp/pouch.json"
	file, err := os.Create(configFile)
	c.Assert(err, check.IsNil)
	defer file.Close()
	defer os.Remove(configFile)

	dcfg := config.Config{
		Debug: true,
	}
	s, _ := json.Marshal(dcfg)
	fmt.Fprintf(file, "%s", s)
	file.Sync()

	daemon.DConfig = daemon.NewConfig()
	daemon.DConfig.Args = append(daemon.DConfig.Args, "--config-file="+configFile)

	//{
	//	err := daemon.DConfig.StartDaemon()
	//	c.Assert(err, check.IsNil)
	//}

	// TODO: verify more

	// Must kill it, as we may loose the pid in next call.
	defer daemon.DConfig.KillDaemon()
}

// TestDaemonInvalideArgs tests invalid args in deamon return error
func (suite *PouchDaemonSuite) TestDaemonInvalideArgs(c *check.C) {
	daemon.DConfig = daemon.NewConfig()
	daemon.DConfig.Args = append(daemon.DConfig.Args, "--config=xxx")

	// depress debug log
	daemon.DConfig.Debug = false

	err := daemon.DConfig.StartDaemon()
	c.Assert(err, check.NotNil)
}

// TestDaemonRestart tests daemon restart with running container.
func (suite *PouchDaemonSuite) TestDaemonRestart(c *check.C) {
	// Start a test daemon with test args.
	daemon.DConfig = daemon.NewConfig()
	err := daemon.DConfig.StartDaemon()
	if err != nil {
		c.Skip("deamon start failed.")
	}
	// Must kill it, as we may loose the pid in next call.
	defer daemon.DConfig.KillDaemon()

	{
		result := command.PouchRun("--host", daemon.Listen, "pull", busyboxImage)
		if result.ExitCode != 0 {
			daemon.DConfig.DumpLog()
			c.Fatalf("pull image failed, err:%v", result)
		}
	}

	cname := "TestDaemonRestart"
	{
		result := command.PouchRun("--host", daemon.Listen, "run", "--name", cname,
			"-p", "1234:80",
			busyboxImage)
		if result.ExitCode != 0 {
			daemon.DConfig.DumpLog()
			c.Fatalf("run container failed, err:%v", result)
		}
	}
	defer DelContainerForceMultyTime(c, cname)

	// restart daemon
	daemon.DConfig.KillDaemon()
	err = daemon.DConfig.StartDaemon()
	c.Assert(err, check.IsNil)

	// test if the container is running.
	output := command.PouchRun("inspect", "--host", daemon.Listen, cname).Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Fatalf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(result.State.Status), check.Equals, "running")
}
