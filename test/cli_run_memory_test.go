package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

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
	SkipIfFalse(c, environment.IsMemorySupport)
	SkipIfFalse(c, environment.IsMemorySwapSupport)

	cname := "TestRunWithMemoryswap"
	m := 1024 * 1024
	memory := "100m"
	memSwap := "200m"
	expected := 200 * m
	sleep := "10000"

	res := command.PouchRun("run", "-d", "-m", memory,
		"--memory-swap", memSwap,
		"--name", cname, busyboxImage, "sleep", sleep)
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.MemorySwap, check.Equals, int64(expected))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf(
		"/sys/fs/cgroup/memory/default/%s/memory.memsw.limit_in_bytes",
		containerID)
	checkFileContains(c, path, strconv.Itoa(expected))

	// test if the value is correct in container
	memSwapLimitFile := "/sys/fs/cgroup/memory/memory.memsw.limit_in_bytes"
	res = command.PouchRun("exec", cname, "cat", memSwapLimitFile)
	res.Assert(c, icmd.Success)

	out := strings.Trim(res.Stdout(), "\n")
	c.Assert(out, check.Equals, strconv.Itoa(expected))

	// test memory swap should be 2x memory if not specify it.
	cname = "TestRunWithMemoryswap2x"
	memory = "10m"
	expected = 2 * 10 * m

	res = command.PouchRun("run", "-d", "-m", memory,
		"--name", cname, busyboxImage, "sleep", sleep)
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result = []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.MemorySwap, check.Equals, int64(expected))

	// test if cgroup has record the real value
	containerID = result[0].ID
	path = fmt.Sprintf(
		"/sys/fs/cgroup/memory/default/%s/memory.memsw.limit_in_bytes",
		containerID)
	checkFileContains(c, path, strconv.Itoa(expected))

	// test if the value is correct in container
	res = command.PouchRun("exec", cname, "cat", memSwapLimitFile)
	res.Assert(c, icmd.Success)

	out = strings.Trim(res.Stdout(), "\n")
	c.Assert(out, check.Equals, strconv.Itoa(expected))
}

// TestRunWithMemoryswappiness is to verify the valid running container
// with memory-swappiness
func (suite *PouchRunMemorySuite) TestRunWithMemoryswappiness(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)
	SkipIfFalse(c, environment.IsMemorySwappinessSupport)

	cname := "TestRunWithMemoryswappiness-1"
	res := command.PouchRun("run", "-d",
		"--memory-swappiness", "-1",
		"--name", cname, busyboxImage, "top")
	DelContainerForceMultyTime(c, cname)
	c.Assert(res.ExitCode, check.Equals, 1)

	cname = "TestRunWithMemoryswappiness"
	memory := "100m"
	memSwappiness := "70"
	expected, _ := strconv.Atoi(memSwappiness)
	sleep := "10000"

	res = command.PouchRun("run", "-d", "-m", memory,
		"--memory-swappiness", memSwappiness,
		"--name", cname, busyboxImage, "sleep", sleep)
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
		check.Equals, int64(expected))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf(
		"/sys/fs/cgroup/memory/default/%s/memory.swappiness", containerID)
	checkFileContains(c, path, memSwappiness)

	// test if the value is correct in container
	memSwappinessFile := "/sys/fs/cgroup/memory/memory.swappiness"
	res = command.PouchRun("exec", cname, "cat", memSwappinessFile)
	res.Assert(c, icmd.Success)

	out := strings.Trim(res.Stdout(), "\n")
	c.Assert(out, check.Equals, memSwappiness)
}

// TestRunWithLimitedMemory is to verify the valid running container with -m
func (suite *PouchRunMemorySuite) TestRunWithLimitedMemory(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)

	cname := "TestRunWithLimitedMemory"
	m := 1024 * 1024
	memory := "100m"
	expected := 100 * m

	res := command.PouchRun("run", "-d", "-m", memory,
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
	c.Assert(result[0].HostConfig.Memory, check.Equals, int64(expected))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf(
		"/sys/fs/cgroup/memory/default/%s/memory.limit_in_bytes", containerID)

	checkFileContains(c, path, strconv.Itoa(expected))

	// test if the value is correct in container
	memLimitFile := "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	res = command.PouchRun("exec", cname, "cat", memLimitFile)
	res.Assert(c, icmd.Success)

	out := strings.Trim(res.Stdout(), "\n")
	c.Assert(out, check.Equals, strconv.Itoa(expected))
}

// TestRunMemoryOOM is to verify return value when a container is OOM.
func (suite *PouchRunMemorySuite) TestRunMemoryOOM(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)

	cname := "TestRunMemoryOOM"
	ret := command.PouchRun("run", "-m", "20m", "--name", cname, busyboxImage, "sh", "-c", "x=a; while true; do x=$x$x$x$x; done")
	defer DelContainerForceMultyTime(c, cname)
	ret.Assert(c, icmd.Expected{ExitCode: 137})
}

// TestRunWithMemoryFlag test pouch run with memory flags
func (suite *PouchRunSuite) TestRunWithMemoryFlag(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)
	SkipIfFalse(c, environment.IsMemorySwapSupport)

	cname := "RunWithOnlyMemorySwap"
	res := command.PouchRun("run", "-d", "--name", cname, "--memory-swap=1g", busyboxImage, "top")
	defer DelContainerForceMultyTime(c, cname)
	c.Assert(util.PartialEqual(res.Stderr(), "You should always set the Memory limit when using Memoryswap limit"), check.IsNil)

	cname = "RunWithMemorySwapLessThanMemory"
	res = command.PouchRun("run", "-d", "--name", cname, "-m=500m", "--memory-swap=50m", busyboxImage, "top")
	defer DelContainerForceMultyTime(c, cname)
	c.Assert(util.PartialEqual(res.Stderr(), "Minimum memoryswap limit should be larger than memory limit"), check.IsNil)
}
