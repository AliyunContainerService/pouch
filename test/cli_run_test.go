package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunSuite is the test suite for run CLI.
type PouchRunSuite struct{}

func init() {
	check.Suite(&PouchRunSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunSuite) TearDownTest(c *check.C) {
}

// TestRun is to verify the correctness of run container with specified name.
func (suite *PouchRunSuite) TestRun(c *check.C) {
	name := "test-run"

	command.PouchRun("run", "-d", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("ps").Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output %s: should contains container %s\n", out, name)
	}
	DelContainerForceMultyTime(c, name)
}

// TestRunPrintHi is to verify run container with executing a command.
func (suite *PouchRunSuite) TestRunPrintHi(c *check.C) {
	name := "test-run-print-hi"

	res := command.PouchRun("run", "--name", name, busyboxImage, "echo", "hi")
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, "hi") {
		c.Fatalf("unexpected output %s expected hi\n", out)
	}
	DelContainerForceMultyTime(c, name)
}

// TestRunPrintHiByImageID is to verify run container with executing a command by image ID.
func (suite *PouchRunSuite) TestRunPrintHiByImageID(c *check.C) {
	name := "test-run-print-hi-by-image-id"

	res := command.PouchRun("images")
	res.Assert(c, icmd.Success)
	imageID := imagesListToKV(res.Combined())[busyboxImage][0]

	res = command.PouchRun("run", "--name", name, imageID, "echo", "hi")
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, "hi") {
		c.Fatalf("unexpected output %s expected hi\n", out)
	}
	DelContainerForceMultyTime(c, name)
}

// TestRunDeviceMapping is to verify --device param when running a container.
func (suite *PouchRunSuite) TestRunDeviceMapping(c *check.C) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		c.Skip("Host does not have /dev/zero")
	}

	name := "test-run-device-mapping"
	testDev := "/dev/testDev"

	res := command.PouchRun("run", "--name", name, "--device", "/dev/zero:"+testDev, busyboxImage, "ls", testDev)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, testDev) {
		c.Fatalf("unexpected output %s expected %s\n", out, testDev)
	}
	DelContainerForceMultyTime(c, name)
}

// TestRunDevicePermissions is to verify --device permissions mode when running a container.
func (suite *PouchRunSuite) TestRunDevicePermissions(c *check.C) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		c.Skip("Host does not have /dev/zero")
	}

	name := "test-run-device-permissions"
	testDev := "/dev/testDev"
	permissions := "crw-rw-rw-"

	res := command.PouchRun("run", "--name", name, "--device", "/dev/zero:"+testDev+":rwm", busyboxImage, "ls", "-l", testDev)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.HasPrefix(out, permissions) {
		c.Fatalf("Output should begin with %s, got %s\n", permissions, out)
	}
	DelContainerForceMultyTime(c, name)
}

// TestRunDeviceInvalidMode is to verify --device wrong mode when running a container.
func (suite *PouchRunSuite) TestRunDeviceInvalidMode(c *check.C) {
	name := "test-run-device-with-wrong-mode"
	wrongMode := "rxm"

	res := command.PouchRun("run", "--name", name, "--device", "/dev/zero:/dev/zero:"+wrongMode, busyboxImage, "ls", "/dev/zero")
	c.Assert(res.Error, check.NotNil)

	expected := "invalid device mode"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s unexpected output %s. \n", expected, out)
	}
}

// TestRunDeviceDirectory is to verify --device with a device directory when running a container.
func (suite *PouchRunSuite) TestRunDeviceDirectory(c *check.C) {
	if _, err := os.Stat("/dev/snd"); err != nil {
		c.Skip("Host does not have direcory /dev/snd")
	}

	name := "test-run-with-directory-device"
	srcDev := "/dev/snd"

	res := command.PouchRun("run", "--name", name, "--device", srcDev+":/dev:rwm", busyboxImage, "ls", "-l", "/dev")
	res.Assert(c, icmd.Success)

	// /dev/snd contans two device: timer, seq
	expected := "timer"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s, got %s\n", expected, out)
	}
	DelContainerForceMultyTime(c, name)
}

// TestRunWithBadDevice is to verify --device with bad device dir when running a container.
func (suite *PouchRunSuite) TestRunDeviceWithBadDevice(c *check.C) {
	name := "test-run-with-bad-device"

	res := command.PouchRun("run", "--name", name, "--device", "/etc", busyboxImage, "ls", "/etc")
	c.Assert(res.Error, check.NotNil)

	expected := "not a device node"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s unexpected output %s. \n", expected, out)
	}
}

// TestRunInWrongWay tries to run create in wrong way.
func (suite *PouchRunSuite) TestRunInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "unknown flag", args: "-a"},

		// TODO: should add the following cases if ready
		// {name: "missing image name", args: ""},
	} {
		res := command.PouchRun("run", tc.args)
		c.Assert(res.Error, check.NotNil, check.Commentf(tc.name))
	}
}

// TestRunEnableLxcfs is to verify run container with lxcfs.
func (suite *PouchRunSuite) TestRunEnableLxcfs(c *check.C) {
	name := "test-run-lxcfs"

	command.PouchRun("run", "-d", "--name", name, "-m", "512M", "--enableLxcfs=true",
		busyboxImage, "sleep", "10000").Assert(c, icmd.Success)

	res := command.PouchRun("exec", name, "head", "-n", "5", "/proc/meminfo")
	res.Assert(c, icmd.Success)

	// the memory should be equal to 512M
	if out := res.Combined(); !strings.Contains(out, "524288 kB") {
		c.Fatalf("upexpected output %v, expected %s\n", res, "524288 kB")
	}
	DelContainerForceMultyTime(c, name)
}

// Comment this flaky test.
// TestRunRestartPolicyAlways is to verify restart policy always works.
//func (suite *PouchRunSuite) TestRunRestartPolicyAlways(c *check.C) {
//	name := "TestRunRestartPolicyAlways"
//
//	command.PouchRun("run", "--name", name, "-d", "--restart=always", busyboxImage, "sh", "-c", "sleep 10000").Assert(c, icmd.Success)
//	command.PouchRun("stop", name).Assert(c, icmd.Success)
//	time.Sleep(5000 * time.Millisecond)
//
//	res := command.PouchRun("ps")
//	res.Assert(c, icmd.Success)
//
//	if out := res.Combined(); !strings.Contains(out, name) {
//		c.Fatalf("expect container %s to be up: %s\n", name, out)
//	}
//	DelContainerForceMultyTime(c,name)
//}

// TestRunRestartPolicyNone is to verify restart policy none works.
func (suite *PouchRunSuite) TestRunRestartPolicyNone(c *check.C) {
	name := "TestRunRestartPolicyNone"

	command.PouchRun("run", "--name", name, "-d", "--restart=no", busyboxImage, "sh", "-c", "sleep 1").Assert(c, icmd.Success)
	time.Sleep(2000 * time.Millisecond)

	res := command.PouchRun("ps")
	res.Assert(c, icmd.Success)

	if out := res.Combined(); strings.Contains(out, name) {
		c.Fatalf("expect container %s to be exited: %s\n", name, out)
	}
	DelContainerForceMultyTime(c, name)
}

// TestRunWithIPCMode is to verify --specific IPC mode when running a container.
// TODO: test container ipc namespace mode.
func (suite *PouchRunSuite) TestRunWithIPCMode(c *check.C) {
	name := "test-run-with-ipc-mode"

	res := command.PouchRun("run", "-d", "--name", name, "--ipc", "host", busyboxImage)
	res.Assert(c, icmd.Success)
	DelContainerForceMultyTime(c, name)
}

// TestRunWithPIDMode is to verify --specific PID mode when running a container.
// TODO: test container pid namespace mode.
func (suite *PouchRunSuite) TestRunWithPIDMode(c *check.C) {
	name := "test-run-with-pid-mode"

	res := command.PouchRun("run", "-d", "--name", name, "--pid", "host", busyboxImage)
	res.Assert(c, icmd.Success)
	DelContainerForceMultyTime(c, name)
}

// TestRunWithUTSMode is to verify --specific UTS mode when running a container.
func (suite *PouchRunSuite) TestRunWithUTSMode(c *check.C) {
	name := "test-run-with-uts-mode"

	res := command.PouchRun("run", "-d", "--name", name, "--uts", "host", busyboxImage)
	res.Assert(c, icmd.Success)
	DelContainerForceMultyTime(c, name)
}

// TestRunWithSysctls is to verify run container with sysctls.
func (suite *PouchRunSuite) TestRunWithSysctls(c *check.C) {
	sysctl := "net.ipv4.ip_forward=1"
	name := "run-sysctl"

	res := command.PouchRun("run", "-d", "--name", name, "--sysctl", sysctl, busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("exec", name, "cat", "/proc/sys/net/ipv4/ip_forward").Stdout()
	if !strings.Contains(output, "1") {
		c.Fatalf("failed to run a container with sysctls: %s", output)
	}
	DelContainerForceMultyTime(c, name)
}

// TestRunWithUser is to verify run container with user.
func (suite *PouchRunSuite) TestRunWithUser(c *check.C) {
	user := "1001"
	name := "run-user"

	res := command.PouchRun("run", "-d", "--name", name, "--user", user, busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("exec", name, "id", "-u").Stdout()
	if !strings.Contains(output, user) {
		c.Fatalf("failed to run a container with user: %s", output)
	}
	DelContainerForceMultyTime(c, name)

	name = "run-user-admin"
	command.PouchRun("run", "-d", "--name", name, busyboxImage).Assert(c, icmd.Success)
	command.PouchRun("exec", name, "adduser", "--disabled-password", "admin").Assert(c, icmd.Success)
	output = command.PouchRun("exec", "-u", "admin", name, "whoami").Stdout()
	if !strings.Contains(output, "admin") {
		c.Errorf("failed to start a container with user: %s", output)
	}
}

// TestRunWithAppArmor is to verify run container with security option AppArmor.
func (suite *PouchRunSuite) TestRunWithAppArmor(c *check.C) {
	appArmor := "apparmor=unconfined"
	name := "run-apparmor"

	res := command.PouchRun("run", "-d", "--name", name, "--security-opt", appArmor, busyboxImage)
	res.Assert(c, icmd.Success)

	// TODO: do the test more strictly with effective AppArmor profile.

	DelContainerForceMultyTime(c, name)
}

// TestRunWithSeccomp is to verify run container with security option seccomp.
func (suite *PouchRunSuite) TestRunWithSeccomp(c *check.C) {
	seccomp := "seccomp=unconfined"
	name := "run-seccomp"

	res := command.PouchRun("run", "-d", "--name", name, "--security-opt", seccomp, busyboxImage)
	res.Assert(c, icmd.Success)

	// TODO: do the test more strictly with effective seccomp profile.

	DelContainerForceMultyTime(c, name)
}

// TestRunWithCapability is to verify run container with capability.
func (suite *PouchRunSuite) TestRunWithCapability(c *check.C) {
	capability := "NET_ADMIN"
	name := "run-capability"

	res := command.PouchRun("run", "--name", name, "--cap-add", capability, busyboxImage, "brctl", "addbr", "foobar")
	res.Assert(c, icmd.Success)
	DelContainerForceMultyTime(c, name)
}

// TestRunWithoutCapability tests running container with --cap-drop
func (suite *PouchRunSuite) TestRunWithoutCapability(c *check.C) {
	capability := "chown"
	name := "run-capability"
	expt := icmd.Expected{
		Err: "Operation not permitted",
	}
	command.PouchRun("run", "--name", name, "--cap-drop", capability, busyboxImage, "chown", "755", "/tmp").Compare(expt)
	DelContainerForceMultyTime(c, name)
}

// TestRunWithPrivilege is to verify run container with privilege.
func (suite *PouchRunSuite) TestRunWithPrivilege(c *check.C) {
	name := "run-privilege"

	res := command.PouchRun("run", "--name", name, "--privileged", busyboxImage, "brctl", "addbr", "foobar")
	res.Assert(c, icmd.Success)
	DelContainerForceMultyTime(c, name)
}

// TestRunWithBlkioWeight is to verify --specific Blkio Weight when running a container.
func (suite *PouchRunSuite) TestRunWithBlkioWeight(c *check.C) {
	name := "test-run-with-blkio-weight"

	res := command.PouchRun("run", "-d", "--name", name, "--blkio-weight", "500", busyboxImage)
	res.Assert(c, icmd.Success)
	DelContainerForceMultyTime(c, name)
}

// TestRunWithLocalVolume is to verify run container with -v volume works.
func (suite *PouchRunSuite) TestRunWithLocalVolume(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	name := funcname

	command.PouchRun("volume", "create", "--name", funcname).Assert(c, icmd.Success)
	command.PouchRun("run", "--name", name, "-v", funcname+":/tmp", busyboxImage, "touch", "/tmp/test").Assert(c, icmd.Success)

	// check the existence of /mnt/local/function/test
	icmd.RunCommand("stat", "/mnt/local/"+funcname+"/test").Assert(c, icmd.Success)

	DelContainerForceMultyTime(c, name)
	command.PouchRun("volume", "remove", funcname).Assert(c, icmd.Success)
}

// checkFileContains checks the content of fname contains expt
func checkFileContains(c *check.C, fname string, expt string) {
	cmdResult := icmd.RunCommand("cat", fname)
	c.Assert(cmdResult.Error, check.IsNil)
	c.Assert(strings.Contains(string(cmdResult.Stdout()), expt), check.Equals, true)
}

// TestRunWithLimitedMemory is to verify the valid running container with -m
func (suite *PouchRunSuite) TestRunWithLimitedMemory(c *check.C) {
	cname := "TestRunWithLimitedMemory"
	command.PouchRun("run", "-d", "-m", "100m", "--name", cname, busyboxImage).Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.Memory, check.Equals, int64(104857600))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.limit_in_bytes", containerID)

	checkFileContains(c, path, "104857600")

	// remove the container
	DelContainerForceMultyTime(c, cname)
}

// TestRunWithMemoryswap is to verify the valid running container with --memory-swap
func (suite *PouchRunSuite) TestRunWithMemoryswap(c *check.C) {
	cname := "TestRunWithMemoryswap"
	command.PouchRun("run", "-d", "-m", "100m", "--memory-swap", "200m",
		"--name", cname, busyboxImage, "sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.MemorySwap, check.Equals, int64(209715200))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.memsw.limit_in_bytes", containerID)
	checkFileContains(c, path, "209715200")

	// remove the container
	DelContainerForceMultyTime(c, cname)
}

// TestRunWithMemoryswappiness is to verify the valid running container with memory-swappiness
func (suite *PouchRunSuite) TestRunWithMemoryswappiness(c *check.C) {
	cname := "TestRunWithMemoryswappiness"
	command.PouchRun("run", "-d", "-m", "100m", "--memory-swappiness", "70",
		"--name", cname, busyboxImage, "sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(int64(*result[0].HostConfig.MemorySwappiness), check.Equals, int64(70))

	// test if cgroup has record the real value
	containerID := result[0].ID
	path := fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.swappiness", containerID)
	checkFileContains(c, path, "70")

	DelContainerForceMultyTime(c, cname)
}

// TestRunWithCPULimit tests CPU related flags.
func (suite *PouchRunSuite) TestRunWithCPULimit(c *check.C) {
	cname := "TestRunWithCPULimit"
	command.PouchRun("run", "-d",
		"--cpuset-cpus", "0",
		"--cpuset-mems", "0",
		"--cpu-share", "1000",
		"--cpu-period", "1000",
		"--cpu-quota", "1000",
		"--name", cname,
		busyboxImage,
		"sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
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
		path := fmt.Sprintf("/sys/fs/cgroup/cpuset/default/%s/cpuset.cpus", containerID)
		checkFileContains(c, path, "0")
	}
	{
		path := fmt.Sprintf("/sys/fs/cgroup/cpuset/default/%s/cpuset.mems", containerID)
		checkFileContains(c, path, "0")
	}
	{
		path := fmt.Sprintf("/sys/fs/cgroup/cpu/default/%s/cpu.shares", containerID)
		checkFileContains(c, path, "1000")
	}
	{
		path := fmt.Sprintf("/sys/fs/cgroup/cpu/default/%s/cpu.cfs_period_us", containerID)
		checkFileContains(c, path, "1000")
	}
	{
		path := fmt.Sprintf("/sys/fs/cgroup/cpu/default/%s/cpu.cfs_quota_us", containerID)
		checkFileContains(c, path, "1000")
	}

	DelContainerForceMultyTime(c, cname)
}

// TestRunBlockIOWeight tests running container with --blkio-weight flag.
func (suite *PouchRunSuite) TestRunBlockIOWeight(c *check.C) {
	cname := "TestRunBlockIOWeight"
	command.PouchRun("run", "-d", "--blkio-weight", "100",
		"--name", cname, busyboxImage, "sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(result[0].HostConfig.BlkioWeight, check.Equals, uint16(100))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf("/sys/fs/cgroup/blkio/default/%s/blkio.weight", containerID)
		checkFileContains(c, path, "100")
	}
	DelContainerForceMultyTime(c, cname)
}

// TestRunBlockIOWeightDevice tests running container with --blkio-weight-device flag.
func (suite *PouchRunSuite) TestRunBlockIOWeightDevice(c *check.C) {
	cname := "TestRunBlockIOWeightDevice"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	command.PouchRun("run", "-d", "--blkio-weight-device", testDisk+":100",
		"--name", cname, busyboxImage, "sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioWeightDevice), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioWeightDevice[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioWeightDevice[0].Weight, check.Equals, uint16(100))

	// test if cgroup has record the real value
	//containerID := result.ID
	//{
	//	path := fmt.Sprintf("/sys/fs/cgroup/blkio/default/%s/blkio.weight_device", containerID)
	//	checkFileContains(c, path, "100")
	//}
	DelContainerForceMultyTime(c, cname)
}

// TestRunDeviceReadBps tests running container with --device-read-bps flag.
func (suite *PouchRunSuite) TestRunDeviceReadBps(c *check.C) {
	cname := "TestRunDeviceReadBps"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	command.PouchRun("run", "-d", "--device-read-bps", testDisk+":1mb",
		"--name", cname, busyboxImage, "sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioDeviceReadBps), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceReadBps[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceReadBps[0].Rate, check.Equals, uint64(1048576))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf("/sys/fs/cgroup/blkio/default/%s/blkio.throttle.read_bps_device", containerID)
		checkFileContains(c, path, "1048576")
	}
	DelContainerForceMultyTime(c, cname)
}

// TestRunDeviceWriteBps tests running container with --device-write-bps flag.
func (suite *PouchRunSuite) TestRunDeviceWriteBps(c *check.C) {
	cname := "TestRunDeviceWriteBps"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	command.PouchRun("run", "-d", "--device-write-bps", testDisk+":1mb",
		"--name", cname, busyboxImage, "sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioDeviceWriteBps), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteBps[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteBps[0].Rate, check.Equals, uint64(1048576))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf("/sys/fs/cgroup/blkio/default/%s/blkio.throttle.write_bps_device", containerID)
		checkFileContains(c, path, "1048576")
	}
	DelContainerForceMultyTime(c, cname)
}

// TestRunDeviceReadIops tests running container with --device-read-iops flag.
func (suite *PouchRunSuite) TestRunDeviceReadIops(c *check.C) {
	cname := "TestRunDeviceReadIops"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	command.PouchRun("run", "-d", "--device-read-iops", testDisk+":1000",
		"--name", cname, busyboxImage, "sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioDeviceReadIOps), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceReadIOps[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceReadIOps[0].Rate, check.Equals, uint64(1000))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf("/sys/fs/cgroup/blkio/default/%s/blkio.throttle.read_iops_device", containerID)
		checkFileContains(c, path, "1000")
	}
	DelContainerForceMultyTime(c, cname)
}

// TestRunDeviceWriteIops tests running container with --device-write-iops flag.
func (suite *PouchRunSuite) TestRunDeviceWriteIops(c *check.C) {
	cname := "TestRunDeviceWriteIops"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	command.PouchRun("run", "-d", "--device-write-iops", testDisk+":1000",
		"--name", cname, busyboxImage, "sleep", "10000").Stdout()

	// test if the value is in inspect result
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioDeviceWriteIOps), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteIOps[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteIOps[0].Rate, check.Equals, uint64(1000))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf("/sys/fs/cgroup/blkio/default/%s/blkio.throttle.write_iops_device", containerID)
		checkFileContains(c, path, "1000")
	}
	DelContainerForceMultyTime(c, cname)
}

//
func (suite *PouchRunSuite) TestRunAlikernelScheLatSwitch(c *check.C) {
	// TODO: as runc has not implemented it, add test later
	SkipIfFalse(c, environment.IsAliKernel)
}

//
func (suite *PouchRunSuite) TestRunAlikernelMemoryWmarkRatio(c *check.C) {
	// TODO: as runc has not implemented it, add test later
	SkipIfFalse(c, environment.IsAliKernel)
}

//
func (suite *PouchRunSuite) TestRunAlikernelMemoryExtra(c *check.C) {
	// TODO: as runc has not implemented it, add test later
	SkipIfFalse(c, environment.IsAliKernel)
}

//
func (suite *PouchRunSuite) TestRunAlikernelMemoryForceEmptyCtl(c *check.C) {
	// TODO: as runc has not implemented it, add test later
	SkipIfFalse(c, environment.IsAliKernel)
}

// TestRunWithHostFileVolume tests binding a host file as a volume into container.
// fixes https://github.com/alibaba/pouch/issues/813
func (suite *PouchRunSuite) TestRunWithHostFileVolume(c *check.C) {
	// first create a file on the host
	filepath := "/tmp/TestRunWithHostFileVolume.md"
	icmd.RunCommand("touch", filepath).Assert(c, icmd.Success)

	cname := "TestRunWithHostFileVolume"
	command.PouchRun("run", "-d", "--name", cname, "-v", fmt.Sprintf("%s:%s", filepath, filepath), busyboxImage).Assert(c, icmd.Success)

	DelContainerForceMultyTime(c, cname)
}

// TestRunWithCgroupParent tests running container with --cgroup-parent.
func (suite *PouchRunSuite) TestRunWithCgroupParent(c *check.C) {
	// cgroup-parent relative path
	testRunWithCgroupParent(c, "pouch", "TestRunWithRelativePathOfCgroupParent")

	// cgroup-parent absolute path
	testRunWithCgroupParent(c, "/pouch/test", "TestRunWithAbsolutePathOfCgroupParent")
}

func testRunWithCgroupParent(c *check.C, cgroupParent, name string) {
	command.PouchRun("run", "-d", "-m", "300M", "--cgroup-parent", cgroupParent, "--name", name, busyboxImage).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result[0].ID

	// this code slice may not robust, but for this test case is enough.
	if strings.HasPrefix(cgroupParent, "/") {
		cgroupParent = cgroupParent[1:]
	}

	if cgroupParent == "" {
		cgroupParent = "default"
	}

	file := "/sys/fs/cgroup/memory/" + cgroupParent + "/" + containerID + "/memory.limit_in_bytes"
	if _, err := os.Stat(file); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", name)
	}

	out, err := exec.Command("cat", file).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "314572800") {
		c.Fatalf("unexpected output %s expected %s\n", string(out), "314572800")
	}

}

// TestRunInvalidCgroupParent checks that a specially-crafted cgroup parent doesn't cause Docker to crash or start modifying /.
func (suite *PouchRunSuite) TestRunInvalidCgroupParent(c *check.C) {
	testRunInvalidCgroupParent(c, "../../../../../../../../SHOULD_NOT_EXIST", "SHOULD_NOT_EXIST", "cgroup-invalid-test")

	testRunInvalidCgroupParent(c, "/../../../../../../../../SHOULD_NOT_EXIST", "/SHOULD_NOT_EXIST", "cgroup-absolute-invalid-test")
}

func testRunInvalidCgroupParent(c *check.C, cgroupParent, cleanCgroupParent, name string) {
	command.PouchRun("run", "-m", "300M", "--cgroup-parent", cgroupParent, "--name", name, busyboxImage, "cat", "/proc/self/cgroup").Assert(c, icmd.Success)

	// We expect "/SHOULD_NOT_EXIST" to not exist. If not, we have a security issue.
	if _, err := os.Stat("/SHOULD_NOT_EXIST"); err == nil || !os.IsNotExist(err) {
		c.Fatalf("SECURITY: --cgroup-parent with ../../ relative paths cause files to be created in the host (this is bad) !!")
	}
}

// TestRunWithDiskQuota tests running container with --disk-quota.
func (suite *PouchRunSuite) TestRunWithDiskQuota(c *check.C) {
	if !environment.IsDiskQuota() {
		c.Skip("Host does not support disk quota")
	}

	ret := command.PouchRun("run", "--disk-quota", "2000m", "--name", "TestRunWithDiskQuota", busyboxImage, "df")
	defer func() {
		command.PouchRun("rm", "-f", "TestRunWithDiskQuota").Assert(c, icmd.Success)
	}()

	ret.Assert(c, icmd.Success)

	out := ret.Combined()

	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "/") && strings.Contains(line, "2048000") {
			found = true
			break
		}
	}

	c.Assert(found, check.Equals, true)
}

// TestRunWithAnnotation is to verify the valid running container with annotation, and verify SpecAnnotation filed has been in inspect output.
func (suite *PouchRunSuite) TestRunWithAnnotation(c *check.C) {
	cname := "TestRunWithAnnotation"
	command.PouchRun("run", "-d", "--annotation", "a=b", "--annotation", "foo=bar", "--name", cname, busyboxImage).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	// kv in map not in order.
	var annotationSlice []string
	for k, v := range result[0].Config.SpecAnnotation {
		annotationSlice = append(annotationSlice, fmt.Sprintf("%s=%s", k, v))
	}
	annotationStr := strings.Join(annotationSlice, " ")

	c.Assert(util.PartialEqual(annotationStr, "a=b"), check.IsNil)
	c.Assert(util.PartialEqual(annotationStr, "foo=bar"), check.IsNil)
}

// TestRunWithExitCode is to verify the valid running container with exit code != 0.
func (suite *PouchRunSuite) TestRunWithExitCode(c *check.C) {
	cname := "TestRunWithExitCode"
	ret := command.PouchRun("run", "--name", cname, busyboxImage, "sh", "-c", "exit 101")
	// test process exit code $? == 101
	ret.Assert(c, icmd.Expected{ExitCode: 101})

	// test container ExitCode == 101
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].State.ExitCode, check.Equals, int64(101))
}

// TestRunWithDiskQuotaRegular tests running container with --disk-quota.
func (suite *PouchRunSuite) TestRunWithDiskQuotaRegular(c *check.C) {
	if !environment.IsDiskQuota() {
		c.Skip("Host does not support disk quota")
	}

	volumeName := "diskquota-volume"
	containerName := "diskquota-regular"

	ret := command.PouchRun("volume", "create", "-n", volumeName, "-o", "size=256m", "-o", "mount=/data/volume")
	defer func() {
		command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
	}()
	ret.Assert(c, icmd.Success)

	ret = command.PouchRun("run",
		"--disk-quota=1024m",
		`--disk-quota=".*=512m"`,
		`--disk-quota="/mnt/mount1=768m"`,
		"-v", "/data/mount1:/mnt/mount1",
		"-v", "/data/mount2:/mnt/mount2",
		"-v", "diskquota-volume:/mnt/mount3",
		"--name", containerName, busyboxImage, "df")
	defer func() {
		command.PouchRun("rm", "-f", containerName).Assert(c, icmd.Success)
	}()
	ret.Assert(c, icmd.Success)

	out := ret.Stdout()

	rootFound := false
	mount1Found := false
	mount2Found := false
	mount3Found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "/") && strings.Contains(line, "1048576") {
			rootFound = true
			continue
		}

		if strings.Contains(line, "/mnt/mount1") && strings.Contains(line, "786432") {
			mount1Found = true
			continue
		}

		if strings.Contains(line, "/mnt/mount2") && strings.Contains(line, "524288") {
			mount2Found = true
			continue
		}

		if strings.Contains(line, "/mnt/mount3") && strings.Contains(line, "262144") {
			mount3Found = true
			continue
		}
	}

	c.Assert(rootFound, check.Equals, true)
	c.Assert(mount1Found, check.Equals, true)
	c.Assert(mount2Found, check.Equals, true)
	c.Assert(mount3Found, check.Equals, true)
}
