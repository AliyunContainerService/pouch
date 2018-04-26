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

	command.PouchRun("run", "-d", "--cpu-share", "20", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

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

	command.PouchRun("update", "--cpu-share", "40", name).Assert(c, icmd.Success)

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

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestUpdateCpuPeriod is to verify the correctness of updating container cpu-period.
func (suite *PouchUpdateSuite) TestUpdateCpuPeriod(c *check.C) {
	name := "update-container-cpu-period"

	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

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

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestUpdateRunningContainer is to verify the correctness of updating a running container.
func (suite *PouchUpdateSuite) TestUpdateRunningContainer(c *check.C) {
	name := "update-running-container"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

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

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestUpdateStoppedContainer is to verify the correctness of updating a stopped container.
func (suite *PouchUpdateSuite) TestUpdateStoppedContainer(c *check.C) {
	name := "update-stopped-container"

	command.PouchRun("create", "-m", "300M", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

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

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestUpdateContainerInvalidValue is to verify the correctness of updating a container with invalid value.
func (suite *PouchUpdateSuite) TestUpdateContainerInvalidValue(c *check.C) {
	name := "update-container-with-invalid-value"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

	res := command.PouchRun("update", "--memory-swappiness", "-2", name)
	c.Assert(res.Error, check.NotNil)

	expectString := "invalid memory swappiness: -2 (its range is -1 or 0-100)"
	if out := res.Combined(); !strings.Contains(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestUpdateContainerWithoutFlag is to verify the correctness of updating a container without any flag.
func (suite *PouchUpdateSuite) TestUpdateContainerWithoutFlag(c *check.C) {
	name := "update-container-without-flag"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

	command.PouchRun("update", name).Assert(c, icmd.Success)

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestUpdateContainerEnv is to verify the correctness of updating env of container.
func (suite *PouchUpdateSuite) TestUpdateContainerEnv(c *check.C) {
	name := "update-container-env"

	command.PouchRun("create", "-m", "300M", "--name", name, busyboxImage).Assert(c, icmd.Success)

	command.PouchRun("update", "--env", "foo=bar", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	if !utils.StringInSlice(result[0].Config.Env, "foo=bar") {
		c.Errorf("expect 'foo=bar' in container env, but got: %v", result[0].Config.Env)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestUpdateRunningContainerEnv is to verify the correctness of updating env of an running container.
func (suite *PouchUpdateSuite) TestUpdateRunningContainerEnv(c *check.C) {
	name := "update-running-container-env"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

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

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)

}

// TestUpdateContainerLabel is to verify the correctness of updating label of container.
func (suite *PouchUpdateSuite) TestUpdateContainerLabel(c *check.C) {
	name := "update-container-label"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

	command.PouchRun("update", "--label", "foo=bar", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	if v, ok := result[0].Config.Labels["foo"]; !ok || v != "bar" {
		c.Errorf("expect 'foo=bar' in Labels, got: %v", result[0].Config.Labels)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}
