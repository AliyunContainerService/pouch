package main

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunCPUSuite is the test suite for run CLI.
type PouchRunCPUSuite struct{}

func init() {
	check.Suite(&PouchRunCPUSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunCPUSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunCPUSuite) TearDownTest(c *check.C) {
}

// TestRunWithCPULimit tests CPU related flags.
func (suite *PouchRunCPUSuite) TestRunWithCPULimit(c *check.C) {
	cname := "TestRunWithCPULimit"
	res := command.PouchRun("run", "-d",
		"--cpuset-cpus", "0",
		"--cpuset-mems", "0",
		"--cpu-share", "1000",
		"--cpu-period", "1000",
		"--cpu-quota", "1000",
		"--name", cname,
		busyboxImage,
		"sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	// check whether the user setting options are in containers' metadata
	c.Assert(result[0].HostConfig.CpusetMems, check.Equals, "0")
	c.Assert(result[0].HostConfig.CPUShares, check.Equals, int64(1000))
	c.Assert(result[0].HostConfig.CpusetCpus, check.Equals, "0")
	c.Assert(result[0].HostConfig.CPUPeriod, check.Equals, int64(1000))
	c.Assert(result[0].HostConfig.CPUQuota, check.Equals, int64(1000))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/cpuset/default/%s/cpuset.cpus", containerID)
		checkFileContains(c, path, "0")
	}
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/cpuset/default/%s/cpuset.mems", containerID)
		checkFileContains(c, path, "0")
	}
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/cpu/default/%s/cpu.shares", containerID)
		checkFileContains(c, path, "1000")
	}
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/cpu/default/%s/cpu.cfs_period_us", containerID)
		checkFileContains(c, path, "1000")
	}
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/cpu/default/%s/cpu.cfs_quota_us", containerID)
		checkFileContains(c, path, "1000")
	}
}
