package main

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/daemon"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
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
		fmt.Printf("start deamon failed with error:%s", err)
		daemon.DConfig.DumpLog()
		c.Skip("deamon start failed.")
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

	// test if cgroup has the right parent path
	//path := fmt.Sprintf("/sys/fs/cgroup/memory/tmp/%s", result.ID)
	//_, err = os.Stat(path)
	//if err != nil {
	//	daemon.DConfig.DumpLog()
	//	c.Fatalf("get cgroup path failed, err:%s", err)
	//}
}
