package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchUpdateSuite is the test suite for update CLI.
type PouchUpdateSuite struct{}

func init() {
	check.Suite(&PouchUpdateSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchUpdateSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchUpdateSuite) TearDownTest(c *check.C) {
}

// TestUpdateContainerNormalOption is to verify the correctness of updating container cpu.
func (suite *PouchUpdateSuite) TestUpdateContainerNormalOption(c *check.C) {
	name := "TestUpdateContainerNormalOption"

	res := command.PouchRun("run", "-d",
		"--name", name,
		// cpu related options
		"--cpu-shares", "1000",
		"--cpu-period", "1000",
		"--cpu-quota", "1000",
		//"--cpuset-cpus", "0",
		//"--cpuset-mems", "0",
		// memory related options
		"--env", "TEST1=foo1",
		"--env", "TEST2=foo2",
		"--env", "TEST3=foo3",
		"-m", "300M",
		busyboxImage,
		"top")

	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	containerID, err := inspectFilter(name, ".ID")
	c.Assert(err, check.IsNil)

	command.PouchRun("update",
		// cpu related update
		"--cpu-shares", "2000",
		"--cpu-period", "1500",
		"--cpu-quota", "1100",
		"--cpuset-cpus", "0",
		"--cpuset-mems", "0",
		// memory related update
		"-m", "500M",
		// env related update
		"--env", "TEST1=bar1", // updating env to new non-empty value
		"--env", "TEST2=", // update env to empty
		"--env", "TEST3", // update to remove the env
		"--env", "TEST4=bar4", // adding a new env
		// label related update
		"--label", "foo=bar",
		name,
	).Assert(c, icmd.Success)

	{
		// test value check about cpushares
		cpuShareFilePath := fmt.Sprintf("/sys/fs/cgroup/cpu/default/%s/cpu.shares", containerID)
		c.Assert(OsStatErr(cpuShareFilePath), check.IsNil)

		out, err := exec.Command("cat", cpuShareFilePath).Output()
		if err != nil {
			c.Fatalf("failed to execute cat command: %v", err)
		}

		if !strings.Contains(string(out), "2000") {
			c.Fatalf("unexpected output %s expected %s\n", string(out), "2000")
		}

		cpuShares, err := inspectFilter(name, ".HostConfig.CPUShares")
		c.Assert(err, check.IsNil)
		c.Assert(cpuShares, check.Equals, "2000")
	}

	{
		// test value check about cpu period
		cpuPeriodFilePath := fmt.Sprintf("/sys/fs/cgroup/cpu/default/%s/cpu.cfs_period_us", containerID)
		c.Assert(OsStatErr(cpuPeriodFilePath), check.IsNil)

		out, err := exec.Command("cat", cpuPeriodFilePath).Output()
		if err != nil {
			c.Fatalf("failed to execute cat command: %v", err)
		}

		if !strings.Contains(string(out), "1500") {
			c.Fatalf("unexpected output %s expected %s\n", string(out), "1500")
		}

		cpuPeriod, err := inspectFilter(name, ".HostConfig.CPUPeriod")
		c.Assert(err, check.IsNil)
		c.Assert(cpuPeriod, check.Equals, "1500")
	}

	{
		// test value check about cpu quota
		cpuQuotaFilePath := fmt.Sprintf("/sys/fs/cgroup/cpu/default/%s/cpu.cfs_quota_us", containerID)
		c.Assert(OsStatErr(cpuQuotaFilePath), check.IsNil)

		out, err := exec.Command("cat", cpuQuotaFilePath).Output()
		if err != nil {
			c.Fatalf("failed to execute cat command: %v", err)
		}

		if !strings.Contains(string(out), "1100") {
			c.Fatalf("unexpected output %s expected %s\n", string(out), "1100")
		}

		cpuQuota, err := inspectFilter(name, ".HostConfig.CPUQuota")
		c.Assert(err, check.IsNil)
		c.Assert(cpuQuota, check.Equals, "1100")
	}

	{
		// test value check about cpusetCPUs
		cpusetCPUsFilePath := fmt.Sprintf("/sys/fs/cgroup/cpuset/default/%s/cpuset.cpus", containerID)
		c.Assert(OsStatErr(cpusetCPUsFilePath), check.IsNil)

		out, err := exec.Command("cat", cpusetCPUsFilePath).Output()
		if err != nil {
			c.Fatalf("failed to execute cat command: %v", err)
		}

		if !strings.Contains(string(out), "0") {
			c.Fatalf("unexpected output %s expected %s\n", string(out), "0")
		}

		cpusetCPUs, err := inspectFilter(name, ".HostConfig.CpusetCpus")
		c.Assert(err, check.IsNil)
		c.Assert(cpusetCPUs, check.Equals, "0")
	}

	{
		// test value check about cpusetMems
		cpusetMemsFilePath := fmt.Sprintf("/sys/fs/cgroup/cpuset/default/%s/cpuset.mems", containerID)
		c.Assert(OsStatErr(cpusetMemsFilePath), check.IsNil)

		out, err := exec.Command("cat", cpusetMemsFilePath).Output()
		if err != nil {
			c.Fatalf("failed to execute cat command: %v", err)
		}

		if !strings.Contains(string(out), "0") {
			c.Fatalf("unexpected output %s expected %s\n", string(out), "0")
		}

		cpusetMems, err := inspectFilter(name, ".HostConfig.CpusetMems")
		c.Assert(err, check.IsNil)
		c.Assert(cpusetMems, check.Equals, "0")
	}

	{
		// test value check about memory limit
		memLimitFilePath := fmt.Sprintf("/sys/fs/cgroup/memory/default/%s/memory.limit_in_bytes", containerID)
		c.Assert(OsStatErr(memLimitFilePath), check.IsNil)

		out, err := exec.Command("cat", memLimitFilePath).Output()
		if err != nil {
			c.Fatalf("failed to execute cat command: %v", err)
		}

		if !strings.Contains(string(out), "524288000") {
			c.Fatalf("unexpected output %s expected %s\n", string(out), "524288000")
		}

		cpuQuota, err := inspectFilter(name, ".HostConfig.Memory")
		c.Assert(err, check.IsNil)
		c.Assert(cpuQuota, check.Equals, "524288000")
	}

	// test value check about env
	{
		envs, err := inspectFilter(name, ".Config.Env")
		c.Assert(err, check.IsNil)

		if !strings.Contains(envs, "TEST1=bar1") {
			c.Fatalf("expect 'TEST1=bar1' in container env, but got: %s", envs)
		}
		if !strings.Contains(envs, "TEST2=") {
			c.Fatalf("expect 'TEST2=' in container env, but got: %s", envs)
		}
		if strings.Contains(envs, "TEST3") {
			c.Fatalf("expect no 'TEST3' in container env since using '--env TEST#' to remove, but got: %s", envs)
		}
		if !strings.Contains(envs, "TEST4=bar4") {
			c.Fatalf("expect 'TEST4=bar4' in container env, but got: %s", envs)
		}

		output := command.PouchRun("exec", name, "env").Stdout()
		if !strings.Contains(output, "TEST1=bar1") {
			c.Fatalf("env TEST1=bar1 should be in container runtime env, while get: %v", output)
		}
		if !strings.Contains(output, "TEST2=") {
			c.Fatalf("env TEST2= should be in container runtime env, while get: %v", output)
		}
		if strings.Contains(output, "TEST3") {
			c.Fatalf("env TEST3 should not be in container runtime env, while get: %v", output)
		}
		if !strings.Contains(output, "TEST4=bar4") {
			c.Fatalf("env TEST4=bar4 should be in container runtime env, while get: %v", output)
		}
	}

	// test value check aboudoct labels
	{
		labels, err := inspectFilter(name, ".Config.Labels")
		c.Assert(err, check.IsNil)

		if !strings.Contains(labels, "foo:bar") {
			c.Fatalf("expect 'foo:bar' in Labels, got: %s", labels)
		}
	}
}

// TestUpdateCpuMemoryFail is to verify the invalid value of updating container cpu and memory related flags will fail.
func (suite *PouchUpdateSuite) TestUpdateCpuMemoryFail(c *check.C) {
	name := "update-container-cpu-memory-period-fail"

	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("update", "--cpu-period", "10", name)
	c.Assert(res.Stderr(), check.NotNil)
	res = command.PouchRun("update", "--cpu-period", "100000000", name)
	c.Assert(res.Stderr(), check.NotNil)
	res = command.PouchRun("update", "--cpu-period", "-1", name)
	c.Assert(res.Stderr(), check.NotNil)
	res = command.PouchRun("update", "--cpu-quota", "1", name)
	c.Assert(res.Stderr(), check.NotNil)
	res = command.PouchRun("update", "-m", "10000", name)
	c.Assert(res.Stderr(), check.NotNil)
}

// TestUpdateStoppedContainer is to verify the correctness of updating a stopped container.
func (suite *PouchUpdateSuite) TestUpdateStoppedContainer(c *check.C) {
	name := "update-stopped-container"

	res := command.PouchRun("create", "-m", "300M", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	containerID, err := inspectFilter(name, ".ID")
	c.Assert(err, check.IsNil)

	command.PouchRun("update", "-m", "500M", name).Assert(c, icmd.Success)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	file := "/sys/fs/cgroup/memory/default/" + containerID + "/memory.limit_in_bytes"
	if _, err := os.Stat(file); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", containerID)
	}

	out, err := exec.Command("cat", file).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "524288000") {
		c.Fatalf("unexpected output %s expected %s\n", string(out), "524288000")
	}

	memory, err := inspectFilter(name, ".HostConfig.Memory")
	c.Assert(err, check.IsNil)
	c.Assert(memory, check.Equals, "524288000")
}

// TestUpdateContainerWithoutFlag is to verify the correctness of updating a container without any flag.
func (suite *PouchUpdateSuite) TestUpdateContainerWithoutFlag(c *check.C) {
	name := "update-container-without-flag"

	res := command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("update", name).Assert(c, icmd.Success)
}

// TestUpdateStoppedContainerEnv is to verify the correctness of updating env of container.
func (suite *PouchUpdateSuite) TestUpdateStoppedContainerEnv(c *check.C) {
	name := "TestUpdateStoppedContainerEnv"

	res := command.PouchRun("create", "-m", "300M", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("update", "--env", "foo=bar", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	if !utils.StringInSlice(result[0].Config.Env, "foo=bar") {
		c.Errorf("expect 'foo=bar' in container env, but got: %v", result[0].Config.Env)
	}
}

// TestUpdateContainerEnvValue is to verify the correctness of updating env's value of container.
func (suite *PouchUpdateSuite) TestUpdateContainerEnvValue(c *check.C) {
	name := "update-container-env-value"

	res := command.PouchRun("run", "-d", "--env", "foo=bar", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("update", "--env", "foo=bar1", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	if !utils.StringInSlice(result[0].Config.Env, "foo=bar1") {
		c.Errorf("expect 'foo=bar' in container env, but got: %v", result[0].Config.Env)
	}

	if utils.StringInSlice(result[0].Config.Env, "foo=bar") {
		c.Errorf("expect change 'foo=bar' to 'foo=bar1', but got: %v", result[0].Config.Env)
	}

	output = command.PouchRun("exec", name, "env").Stdout()
	if !strings.Contains(output, "foo=bar1") {
		c.Fatalf("Update running container env not worked")
	}
}

// TestUpdateContainerDeleteEnv is to verify the correctness of delete env by update interface
func (suite *PouchUpdateSuite) TestUpdateContainerDeleteEnv(c *check.C) {
	name := "update-container-delete-env"

	res := command.PouchRun("run", "-d", "--env", "foo=bar", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("update", "--env", "foo=", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	if utils.StringInSlice(result[0].Config.Env, "foo=bar") {
		c.Errorf("expect 'foo=bar' env being deleted, but not")
	}

	output = command.PouchRun("exec", name, "env").Stdout()
	if strings.Contains(output, "foo=bar") {
		c.Errorf("foo=bar env should be deleted from container's env")
	}
}

// TestUpdateContainerDiskQuota is to verify the correctness of disk quota by update interface
func (suite *PouchUpdateSuite) TestUpdateContainerDiskQuota(c *check.C) {
	if !environment.IsDiskQuota() {
		c.Skip("Host does not support disk quota")
	}

	// create container with disk quota
	name := "update-container-diskquota"
	command.PouchRun("run", "-d", "--disk-quota", "/=2000m", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	ret := command.PouchRun("exec", name, "df")
	//ret.Assert(c, icmd.Success)
	out := ret.Combined()

	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "/") &&
			strings.Contains(line, "2048000") {
			found = true
			break
		}
	}
	c.Assert(found, check.Equals, true)

	// update diskquota
	command.PouchRun("update", "--disk-quota", "/=1000m", name).Assert(c, icmd.Success)

	ret = command.PouchRun("exec", name, "df")
	//ret.Assert(c, icmd.Success)
	out = ret.Combined()

	found = false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "/") &&
			strings.Contains(line, "1024000") {
			found = true
			break
		}
	}
	c.Assert(found, check.Equals, true)
}

func checkContainerCPUQuota(c *check.C, cName, cpuQuota string) {
	var (
		containerID    string
		cgroupCPUQuota = cpuQuota
	)

	output := command.PouchRun("inspect", cName).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID = result[0].ID

	if strconv.FormatInt(result[0].HostConfig.CPUQuota, 10) != cpuQuota {
		c.Errorf("expect CPUQuota %s, but got: %v", cpuQuota, result[0].HostConfig.CPUQuota)
	}

	// container's cpu-quota default is 0 that means not limit cpu quota in cgroup, and
	// cpu.cfs_quota_us value is -1 when not limit cpu quota in cgroup.
	if cgroupCPUQuota == "0" {
		cgroupCPUQuota = "-1"
	}
	path := fmt.Sprintf("/sys/fs/cgroup/cpu/default/%s/cpu.cfs_quota_us", containerID)
	checkFileContains(c, path, cgroupCPUQuota)
}

// TestUpdateContainerCPUQuota is to verify the correctness of update cpuquota by update interface
func (suite *PouchUpdateSuite) TestUpdateContainerCPUQuota(c *check.C) {
	name := "TestUpdateContainerCPUQuota"

	command.PouchRun("run", "-d",
		"--name", name,
		busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// default cpuquota should be 0
	checkContainerCPUQuota(c, name, "0")

	// update cpuquota to 0, should not take effect
	command.PouchRun("update", "--cpu-quota", "0", name).Assert(c, icmd.Success)
	// 0 is a meaningless value
	checkContainerCPUQuota(c, name, "0")

	// update not specified any parameters,  cpuquota should still be 0
	command.PouchRun("update", name).Assert(c, icmd.Success)
	checkContainerCPUQuota(c, name, "0")

	// update cpuquota to [1, 1000), should return error
	res := command.PouchRun("update", "--cpu-quota", "20", name)
	c.Assert(res.Stderr(), check.NotNil, check.Commentf("CPU cfs quota should be greater than 1ms(1000)"))

	// update cpuquota to 1100, should take effect
	command.PouchRun("update", "--cpu-quota", "1100", name).Assert(c, icmd.Success)
	checkContainerCPUQuota(c, name, "1100")

	// update cpuquota to -1, should take effect
	command.PouchRun("update", "--cpu-quota", "-1", name).Assert(c, icmd.Success)
	checkContainerCPUQuota(c, name, "-1")

}

// TestUpdateStoppedContainerCPUQuota is to verify the correctness of update the cpuquota
// of a stopped container by update interface
func (suite *PouchUpdateSuite) TestUpdateStoppedContainerCPUQuota(c *check.C) {
	name := "TestUpdateContainerCPUQuota"

	command.PouchRun("create",
		"--cpu-quota", "1100",
		"--name", name,
		busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// update cpuquota to 1200, should take effect
	command.PouchRun("update", "--cpu-quota", "1200", name).Assert(c, icmd.Success)

	// start container
	command.PouchRun("start", name).Assert(c, icmd.Success)

	// then check the cpu-quota value
	checkContainerCPUQuota(c, name, "1200")

	// update cpuquota to 0, should not take effect
	command.PouchRun("update", "--cpu-quota", "0", name).Assert(c, icmd.Success)
	// 0 is a meaningless value
	checkContainerCPUQuota(c, name, "1200")

	// update not specified any parameters,  cpuquota should still be 1200
	command.PouchRun("update", name).Assert(c, icmd.Success)
	checkContainerCPUQuota(c, name, "1200")

}

// TestUpdateBlkIOLimit is to verify the correctness of update read/write bps/iops
func (suite *PouchUpdateSuite) TestUpdateBlkIOLimit(c *check.C) {
	cname := "TestUpdateBlkIOLimit"
	testDisk := "/dev/null"

	number, exist := util.GetMajMinNumOfDevice(testDisk)
	if !exist {
		c.Skip("fail to get major:minor device number")
	}

	oldReadBpsDev := testDisk + ":1000"
	oldReadIopsDev := testDisk + ":1500"
	oldWriteBpsDev := testDisk + ":2000"
	oldWriteIopsDev := testDisk + ":2500"
	newReadBpsDev := testDisk + ":3000"
	newReadIopsDev := testDisk + ":3500"
	newWriteBpsDev := testDisk + ":4000"
	newWriteIopsDev := testDisk + ":4500"

	blkioDeviceReadBpsFile := "/sys/fs/cgroup/blkio/blkio.throttle.read_bps_device"
	blkioDeviceReadIopsFile := "/sys/fs/cgroup/blkio/blkio.throttle.read_iops_device"
	blkioDeviceWriteBpsFile := "/sys/fs/cgroup/blkio/blkio.throttle.write_bps_device"
	blkioDeviceWriteIopsFile := "/sys/fs/cgroup/blkio/blkio.throttle.write_iops_device"

	Expected := fmt.Sprintf("%s %s %s %s %s %s %s %s\n", number, "3000", number, "3500", number, "4000", number, "4500")

	res := command.PouchRun("run", "-d", "--name", cname, "--device-read-bps", oldReadBpsDev, "--device-read-iops", oldReadIopsDev,
		"--device-write-bps", oldWriteBpsDev, "--device-write-iops", oldWriteIopsDev, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	command.PouchRun("update", "--device-read-bps", newReadBpsDev, cname).Assert(c, icmd.Success)
	command.PouchRun("update", "--device-read-iops", newReadIopsDev, cname).Assert(c, icmd.Success)
	command.PouchRun("update", "--device-write-bps", newWriteBpsDev, cname).Assert(c, icmd.Success)
	command.PouchRun("update", "--device-write-iops", newWriteIopsDev, cname).Assert(c, icmd.Success)

	// Using "sed" to convert output to one line
	res = command.PouchRun("exec", cname, "cat", blkioDeviceReadBpsFile, blkioDeviceReadIopsFile, blkioDeviceWriteBpsFile,
		blkioDeviceWriteIopsFile, "| sed ':a;N;s/\n/ /g;ta'")
	out := res.Stdout()
	c.Assert(out, check.Equals, Expected)
}

func checkContainerAnnotation(c *check.C, cName string, annotationKey string, expect string) {
	output := command.PouchRun("inspect", cName).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	annotations := result[0].Config.SpecAnnotation
	v, found := annotations[annotationKey]
	c.Assert(found, check.Equals, true)
	c.Assert(v, check.Equals, expect)
}

// TestUpdateAnnotation is to verity the correctness of update the annotation
func (suite *PouchUpdateSuite) TestUpdateAnnotation(c *check.C) {
	cname := "TestUpdateAnnotation1"
	annotation1 := "key1=value1"
	annotation2 := "key2=value2"
	command.PouchRun("run", "-d", "--name", cname, "--annotation", annotation1, "--annotation", annotation2, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	annotation1Update := "key1=value1.new"
	annotation2Update := "key2=value2,value3"

	command.PouchRun("update", "--annotation", annotation1Update, cname).Assert(c, icmd.Success)
	checkContainerAnnotation(c, cname, "key1", "value1.new")
	checkContainerAnnotation(c, cname, "key2", "value2")

	command.PouchRun("update", "--annotation", annotation2Update, cname).Assert(c, icmd.Success)
	checkContainerAnnotation(c, cname, "key1", "value1.new")
	checkContainerAnnotation(c, cname, "key2", "value2,value3")

	command.PouchRun("restart", cname).Assert(c, icmd.Success)
	checkContainerAnnotation(c, cname, "key1", "value1.new")
	checkContainerAnnotation(c, cname, "key2", "value2,value3")
}
