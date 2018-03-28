package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchUpgradeSuite is the test suite for upgrade CLI.
type PouchUpgradeSuite struct{}

func init() {
	check.Suite(&PouchUpgradeSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchUpgradeSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
	command.PouchRun("pull", busyboxImage125).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchUpgradeSuite) TeadDownTest(c *check.C) {
	// TODO
}

// TestPouchUpgrade is to verify pouch upgrade command.
func (suite *PouchUpgradeSuite) TestPouchUpgrade(c *check.C) {
	name := "TestPouchUpgrade"

	command.PouchRun("run", "-d", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("upgrade", "--name", name, busyboxImage125)
	c.Assert(res.Error, check.IsNil)

	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output: %s, expected: %s", out, name)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)

}

// TestPouchUpgradeNoChange is to verify pouch upgrade command with same image.
func (suite *PouchUpgradeSuite) TestPouchUpgradeNoChange(c *check.C) {
	name := "TestPouchUpgradeNoChange"

	command.PouchRun("run", "-d", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("upgrade", "--name", name, busyboxImage)
	c.Assert(res.Error, check.NotNil)

	expectedStr := "failed to upgrade container: image not changed"
	if out := res.Combined(); !strings.Contains(out, expectedStr) {
		c.Fatalf("unexpected output: %s, expected: %s", out, expectedStr)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)

}

// TestPouchUpgradeStoppedContainer is to verify pouch upgrade a stopped command.
func (suite *PouchUpgradeSuite) TestPouchUpgradeStoppedContainer(c *check.C) {
	name := "TestPouchUpgradeStoppedContainer"

	command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("upgrade", "--name", name, busyboxImage125)
	c.Assert(res.Error, check.IsNil)

	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output: %s, expected %s", out, name)
	}

	command.PouchRun("start", name).Assert(c, icmd.Success)

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestPouchUpgradeContainerMemCpu is to verify pouch upgrade container's memory
func (suite *PouchUpgradeSuite) TestPouchUpgradeContainerMemCpu(c *check.C) {
	name := "TestPouchUpgradeContainerMemCpu"

	command.PouchRun("run", "-d", "-m", "300m", "--cpu-share", "20", "--name", name, busyboxImage).Assert(c, icmd.Success)

	command.PouchRun("upgrade", "-m", "500m", "--cpu-share", "40", "--name", name, busyboxImage125).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result.ID

	// Check if metajson has changed
	c.Assert(result.HostConfig.Memory, check.Equals, int64(524288000))
	c.Assert(result.HostConfig.CPUShares, check.Equals, int64(40))

	// Check if cgroup file has changed
	memFile := "/sys/fs/cgroup/memory/default/" + containerID + "/memory.limit_in_bytes"
	if _, err := os.Stat(memFile); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", containerID)
	}

	out, err := exec.Command("cat", memFile).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "524288000") {
		c.Fatalf("unexpected output %s expected %s\n", string(out), "524288000")
	}

	cpuFile := "/sys/fs/cgroup/cpu/default/" + containerID + "/cpu.shares"
	if _, err := os.Stat(cpuFile); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", containerID)
	}

	out, err = exec.Command("cat", cpuFile).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "40") {
		c.Fatalf("unexpected output %s expected %s\n", string(out), "40")
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestPouchUpgradeContainerLabels is to verify pouch upgrade container's labels
func (suite *PouchUpgradeSuite) TestPouchUpgradeContainerLabels(c *check.C) {
	name := "TestPouchUpgradeContainerLabels"

	command.PouchRun("run", "-d", "--label", "test=foo", "--name", name, busyboxImage).Assert(c, icmd.Success)

	command.PouchRun("upgrade", "--label", "test1=bar", "--name", name, busyboxImage125).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	labels := map[string]string{
		"test":  "foo",
		"test1": "bar",
	}

	if !reflect.DeepEqual(result.Config.Labels, labels) {
		c.Errorf("unexpected output: %s, expected: %s", result.Config.Labels, labels)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}
