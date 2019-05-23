package main

import (
	"fmt"
	"strconv"
	"strings"

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
func (suite *PouchRunMemorySuite) TestRunWithMemoryswapAndKernelMemory(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)
	SkipIfFalse(c, environment.IsMemorySwapSupport)

	cname := "TestRunWithMemoryswapAndKernelMemory"
	memory := "100m"
	memSwap := "200m"
	kernelMemory := "100m"
	expectedMem := strconv.Itoa(100 * 1024 * 1024)       // 100 MB
	expectedMemSwap := strconv.Itoa(200 * 1024 * 1024)   // 200 MB
	expectedKernelMem := strconv.Itoa(1024 * 1024 * 100) // 100 MB

	res := command.PouchRun("run", "-d",
		"--memory", memory,
		"--memory-swap", memSwap,
		"--kernel-memory", kernelMemory,
		"--name", cname,
		busyboxImage,
		"sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	memoryResult, err := inspectFilter(cname, ".HostConfig.Memory")
	c.Assert(err, check.IsNil)
	c.Assert(memoryResult, check.Equals, expectedMem)
	memorySwapResult, err := inspectFilter(cname, ".HostConfig.MemorySwap")
	c.Assert(err, check.IsNil)
	c.Assert(memorySwapResult, check.Equals, expectedMemSwap)
	kernelMemoryResult, err := inspectFilter(cname, ".HostConfig.KernelMemory")
	c.Assert(err, check.IsNil)
	c.Assert(kernelMemoryResult, check.Equals, expectedKernelMem)

	// test if cgroup has record the real value
	containerID, err := inspectFilter(cname, ".ID")
	c.Assert(err, check.IsNil)
	path := fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.limit_in_bytes", containerID)
	checkFileContains(c, path, expectedMem)
	path = fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.memsw.limit_in_bytes", containerID)
	checkFileContains(c, path, expectedMemSwap)
	path = fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.kmem.limit_in_bytes", containerID)
	checkFileContains(c, path, expectedKernelMem)

	// test if the value is correct in container
	memSwapLimitFile := "/sys/fs/cgroup/memory/memory.memsw.limit_in_bytes"
	res = command.PouchRun("exec", cname, "cat", memSwapLimitFile)
	res.Assert(c, icmd.Success)
	out := strings.Trim(res.Stdout(), "\n")
	c.Assert(out, check.Equals, expectedMemSwap)
}

// test memory swap should be 2x memory if not specify it.
func (suite *PouchRunMemorySuite) TestRunWithMemoryswap2x(c *check.C) {
	cname := "TestRunWithMemoryswap2x"
	memory := "20m"
	expectedMemorySwap := strconv.Itoa(2 * 20 * 1024 * 1024) // 40 MB

	res := command.PouchRun("run", "-d",
		"-m", memory,
		"--name", cname,
		busyboxImage,
		"sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	memorySwap, err := inspectFilter(cname, ".HostConfig.MemorySwap")
	c.Assert(err, check.IsNil)
	c.Assert(memorySwap, check.Equals, expectedMemorySwap)

	// test if cgroup has record the real value
	containerID, err := inspectFilter(cname, ".ID")
	c.Assert(err, check.IsNil)
	path := fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.memsw.limit_in_bytes", containerID)
	checkFileContains(c, path, expectedMemorySwap)

	// test if the value is correct in container
	memSwapLimitFile := "/sys/fs/cgroup/memory/memory.memsw.limit_in_bytes"
	res = command.PouchRun("exec", cname, "cat", memSwapLimitFile)
	res.Assert(c, icmd.Success)
	out := strings.Trim(res.Stdout(), "\n")
	c.Assert(out, check.Equals, expectedMemorySwap)
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
	c.Assert(res.ExitCode, check.Equals, 0)

	cname = "TestRunWithMemoryswappiness"
	memory := "100m"
	memSwappiness := "70"
	sleep := "10000"

	res = command.PouchRun("run", "-d", "-m", memory,
		"--memory-swappiness", memSwappiness,
		"--name", cname, busyboxImage, "sleep", sleep)
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	memorySwappiness, err := inspectFilter(cname, ".HostConfig.MemorySwappiness")
	c.Assert(err, check.IsNil)
	c.Assert(memorySwappiness, check.Equals, memSwappiness)

	// test if cgroup has record the real value
	containerID, err := inspectFilter(cname, ".ID")
	c.Assert(err, check.IsNil)
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

// TestRunWithMemoryReservation is to verify the valid running container
// with memory-reservation
func (suite *PouchRunMemorySuite) TestRunWithMemoryReservation(c *check.C) {
	SkipIfFalse(c, environment.IsMemoryReservationSupport)

	// test run with invalid memory reservation
	cname := "TestRunWithMemoryReservationInvalid"
	res := command.PouchRun("run", "-d",
		"-m", "5m",
		"--memory-reservation", "8m",
		"--name", cname, busyboxImage, "true")
	DelContainerForceMultyTime(c, cname)
	c.Assert(res.ExitCode, check.Not(check.Equals), 0)

	cname = "TestRunWithMemoryReservationInvalid2"
	res = command.PouchRun("run", "-d",
		"--memory-reservation", "1k",
		"--name", cname, busyboxImage, "true")
	DelContainerForceMultyTime(c, cname)
	c.Assert(res.ExitCode, check.Not(check.Equals), 0)

	// test run with memory reservation
	cname = "TestRunWithMemoryReservation"
	memReservation := "200m"
	expected := "209715200"

	res = command.PouchRun("run", "-d",
		"--memory-reservation", memReservation,
		"--name", cname,
		busyboxImage,
		"sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	memoryReservationResult, err := inspectFilter(cname, ".HostConfig.MemoryReservation")
	c.Assert(err, check.IsNil)
	c.Assert(memoryReservationResult, check.Equals, expected)

	// test if cgroup has record the real value
	containerID, err := inspectFilter(cname, ".ID")
	c.Assert(err, check.IsNil)
	path := fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.soft_limit_in_bytes", containerID)
	checkFileContains(c, path, expected)

	// test if the value is correct in container
	memReservationFile := "/sys/fs/cgroup/memory/memory.soft_limit_in_bytes"
	res = command.PouchRun("exec", cname, "cat", memReservationFile)
	res.Assert(c, icmd.Success)
	out := strings.Trim(res.Stdout(), "\n")
	c.Assert(out, check.Equals, expected)
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
	memory, err := inspectFilter(cname, ".HostConfig.Memory")
	c.Assert(err, check.IsNil)
	c.Assert(memory, check.Equals, strconv.Itoa(expected))

	// test if cgroup has record the real value
	containerID, err := inspectFilter(cname, ".ID")
	c.Assert(err, check.IsNil)
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
func (suite *PouchRunMemorySuite) TestRunWithMemoryFlag(c *check.C) {
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

// TestRunWithShm is to verify the valid running container
// with shm-size
func (suite *PouchRunMemorySuite) TestRunWithShm(c *check.C) {
	cname := "TestRunWithShm"
	res := command.PouchRun("run", "-d", "--shm-size", "1g",
		"--name", cname, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	shmSize, err := inspectFilter(cname, ".HostConfig.ShmSize")
	c.Assert(err, check.IsNil)
	c.Assert(shmSize,
		check.Equals, "1073741824")

	res = command.PouchRun("exec", cname, "df", "-k", "/dev/shm")
	res.Assert(c, icmd.Success)

	c.Assert(util.PartialEqual(res.Stdout(), "1048576"), check.IsNil)
}

// TestRunWithShm is to verify the container has default shm size equals 64m
func (suite *PouchRunMemorySuite) TestRunWithDefaultShm(c *check.C) {
	cname := "TestRunWithDefaultShm"
	res := command.PouchRun("run", "-d", "--name", cname, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("exec", cname, "df", "-k", "/dev/shm")
	res.Assert(c, icmd.Success)

	c.Assert(util.PartialEqual(res.Stdout(), "65536"), check.IsNil)
}
