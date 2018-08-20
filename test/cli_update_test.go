package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

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

// TestUpdateCpu is to verify the correctness of updating container cpu.
func (suite *PouchUpdateSuite) TestUpdateCpu(c *check.C) {
	name := "update-container-cpu"

	res := command.PouchRun("run", "-d", "--cpu-shares", "20", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result[0].ID

	file := "/sys/fs/cgroup/cpu/default/" + containerID + "/cpu.shares"
	if _, err := os.Stat(file); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", containerID)
	}

	command.PouchRun("update", "--cpu-shares", "40", name).Assert(c, icmd.Success)

	out, err := exec.Command("cat", file).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "40") {
		c.Fatalf("unexpected output %s expected %s\n", string(out), "40")
	}

	inspectInfo := command.PouchRun("inspect", name).Stdout()
	metaJSON := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(inspectInfo), &metaJSON); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(metaJSON[0].HostConfig.CPUShares, check.Equals, int64(40))
}

// TestUpdateCpuPeriod is to verify the correctness of updating container cpu-period.
func (suite *PouchUpdateSuite) TestUpdateCpuPeriod(c *check.C) {
	name := "update-container-cpu-period"

	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result[0].ID

	file := "/sys/fs/cgroup/cpu/default/" + containerID + "/cpu.cfs_period_us"
	if _, err := os.Stat(file); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", containerID)
	}

	command.PouchRun("update", "--cpu-period", "2000", name).Assert(c, icmd.Success)

	out, err := exec.Command("cat", file).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "2000") {
		c.Fatalf("unexpected output %s expected %s\n", string(out), "2000")
	}

	inspectInfo := command.PouchRun("inspect", name).Stdout()
	metaJSON := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(inspectInfo), &metaJSON); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(metaJSON[0].HostConfig.CPUPeriod, check.Equals, int64(2000))
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

// TestUpdateRunningContainer is to verify the correctness of updating a running container.
func (suite *PouchUpdateSuite) TestUpdateRunningContainer(c *check.C) {
	name := "update-running-container"

	res := command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result[0].ID

	file := "/sys/fs/cgroup/memory/default/" + containerID + "/memory.limit_in_bytes"
	if _, err := os.Stat(file); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", containerID)
	}

	command.PouchRun("update", "-m", "500M", name).Assert(c, icmd.Success)

	out, err := exec.Command("cat", file).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "524288000") {
		c.Fatalf("unexpected output %s expected %s\n", string(out), "524288000")
	}

	inspectInfo := command.PouchRun("inspect", name).Stdout()
	metaJSON := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(inspectInfo), &metaJSON); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(metaJSON[0].HostConfig.Memory, check.Equals, int64(524288000))
}

// TestUpdateStoppedContainer is to verify the correctness of updating a stopped container.
func (suite *PouchUpdateSuite) TestUpdateStoppedContainer(c *check.C) {
	name := "update-stopped-container"

	res := command.PouchRun("create", "-m", "300M", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result[0].ID

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

	inspectInfo := command.PouchRun("inspect", name).Stdout()
	metaJSON := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(inspectInfo), &metaJSON); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(metaJSON[0].HostConfig.Memory, check.Equals, int64(524288000))
}

// TestUpdateContainerCPUQuota is to verify the correctness of updating cpu-quota of container.
func (suite *PouchUpdateSuite) TestUpdateContainerCPUQuota(c *check.C) {
	name := "update-container-cpu-quota"

	res := command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	// ensure update cpu-quota is ok
	command.PouchRun("update", "--cpu-quota", "1100", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result[0].ID

	file := "/sys/fs/cgroup/cpu/default/" + containerID + "/cpu.cfs_quota_us"
	if _, err := os.Stat(file); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", containerID)
	}

	out, err := exec.Command("cat", file).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "1100") {
		c.Fatalf("unexpected output %s expected %s\n", string(out), "524288000")
	}

	inspectInfo := command.PouchRun("inspect", name).Stdout()
	metaJSON := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(inspectInfo), &metaJSON); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(metaJSON[0].HostConfig.CPUQuota, check.Equals, int64(1100))
}

// TestUpdateContainerWithoutFlag is to verify the correctness of updating a container without any flag.
func (suite *PouchUpdateSuite) TestUpdateContainerWithoutFlag(c *check.C) {
	name := "update-container-without-flag"

	res := command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("update", name).Assert(c, icmd.Success)
}

// TestUpdateContainerEnv is to verify the correctness of updating env of container.
func (suite *PouchUpdateSuite) TestUpdateContainerEnv(c *check.C) {
	name := "update-container-env"

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

// TestUpdateRunningContainerEnv is to verify the correctness of updating env of an running container.
func (suite *PouchUpdateSuite) TestUpdateRunningContainerEnv(c *check.C) {
	name := "update-running-container-env"

	res := command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top")
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

	output = command.PouchRun("exec", name, "env").Stdout()
	if !strings.Contains(output, "foo=bar") {
		c.Fatalf("Update running container env not worked")
	}
}

// TestUpdateContainerLabel is to verify the correctness of updating label of container.
func (suite *PouchUpdateSuite) TestUpdateContainerLabel(c *check.C) {
	name := "update-container-label"

	res := command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("update", "--label", "foo=bar", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	if v, ok := result[0].Config.Labels["foo"]; !ok || v != "bar" {
		c.Errorf("expect 'foo=bar' in Labels, got: %v", result[0].Config.Labels)
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

// TestUpdateContainerDiskQuota is to verify the correctness of delete env by update interface
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
