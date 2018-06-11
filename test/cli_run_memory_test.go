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

// PouchRunMemorySuite is the test suite for run CLI.
type PouchRunMemorySuite struct{}

func init() {
	check.Suite(&PouchRunMemorySuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunMemorySuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunMemorySuite) TearDownTest(c *check.C) {
}

// TestRunWithMemoryswap is to verify the valid running container
// with --memory-swap
func (suite *PouchRunMemorySuite) TestRunWithMemoryswap(c *check.C) {
	cname := "TestRunWithMemoryswap"
	res := command.PouchRun("run", "-d", "-m", "100m",
		"--memory-swap", "200m",
		"--name", cname, busyboxImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.MemorySwap, check.Equals, int64(209715200))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf(
		"/sys/fs/cgroup/memory/default/%s/memory.memsw.limit_in_bytes",
		containerID)
	checkFileContains(c, path, "209715200")

	// test memory swap should be 2x memory if not specify it.
	cname = "TestRunWithMemoryswap2x"
	res = command.PouchRun("run", "-d", "-m", "10m",
		"--name", cname, busyboxImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result = []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.MemorySwap, check.Equals, int64(20971520))

	// test if cgroup has record the real value
	containerID = result[0].ID
	path = fmt.Sprintf(
		"/sys/fs/cgroup/memory/default/%s/memory.memsw.limit_in_bytes",
		containerID)
	checkFileContains(c, path, "20971520")
}

// TestRunWithMemoryswappiness is to verify the valid running container
// with memory-swappiness
func (suite *PouchRunMemorySuite) TestRunWithMemoryswappiness(c *check.C) {
	cname := "TestRunWithMemoryswappiness"
	res := command.PouchRun("run", "-d", "-m", "100m",
		"--memory-swappiness", "70",
		"--name", cname, busyboxImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(int64(*result[0].HostConfig.MemorySwappiness),
		check.Equals, int64(70))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf(
		"/sys/fs/cgroup/memory/default/%s/memory.swappiness", containerID)
	checkFileContains(c, path, "70")
}

// TestRunWithLimitedMemory is to verify the valid running container with -m
func (suite *PouchRunMemorySuite) TestRunWithLimitedMemory(c *check.C) {
	cname := "TestRunWithLimitedMemory"
	res := command.PouchRun("run", "-d", "-m", "100m",
		"--name", cname, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.Memory, check.Equals, int64(104857600))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf(
		"/sys/fs/cgroup/memory/default/%s/memory.limit_in_bytes", containerID)

	checkFileContains(c, path, "104857600")
}
