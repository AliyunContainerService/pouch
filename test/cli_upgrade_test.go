package main

import (
	"encoding/json"
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

	PullImage(c, busyboxImage)
	PullImage(c, busyboxImage125)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchUpgradeSuite) TeadDownTest(c *check.C) {
	// TODO
}

// TestPouchUpgrade is to verify pouch upgrade command.
func (suite *PouchUpgradeSuite) TestPouchUpgrade(c *check.C) {
	name := "TestPouchUpgrade"

	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("upgrade", "--image", busyboxImage125, name)
	res.Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output: %s, expected: %s", out, name)
	}

	// check if the new container is running after upgade a running container
	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].State.Running, check.Equals, true)

	// double check if container is running by executing a exec command
	out := command.PouchRun("exec", name, "echo", "test").Stdout()
	if !strings.Contains(out, "test") {
		c.Errorf("failed to exec in container, expected test got %s", out)
	}
}

// TestPouchUpgradeNoChange is to verify pouch upgrade command with same image.
func (suite *PouchUpgradeSuite) TestPouchUpgradeNoChange(c *check.C) {
	name := "TestPouchUpgradeNoChange"

	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("upgrade", "--image", busyboxImage, name)
	c.Assert(res.Stderr(), check.NotNil)

	expectedStr := "failed to upgrade container: image not changed"
	if out := res.Combined(); !strings.Contains(out, expectedStr) {
		c.Fatalf("unexpected output: %s, expected: %s", out, expectedStr)
	}
}

// TestPouchUpgradeStoppedContainer is to verify pouch upgrade a stopped command.
func (suite *PouchUpgradeSuite) TestPouchUpgradeStoppedContainer(c *check.C) {
	name := "TestPouchUpgradeStoppedContainer"

	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("stop", name).Assert(c, icmd.Success)

	res := command.PouchRun("upgrade", "--image", busyboxImage125, name)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output: %s, expected %s", out, name)
	}

	// check if the new container is running after upgade a running container
	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].State.Status, check.Equals, types.StatusStopped)

	command.PouchRun("start", name).Assert(c, icmd.Success)
}

// TestPouchUpgradeWithDifferentImage is to verify pouch upgrade command.
func (suite *PouchUpgradeSuite) TestPouchUpgradeWithDifferentImage(c *check.C) {
	name := "TestPouchUpgradeWithDifferentImage"
	command.PouchRun("run", "-d", "--name", name, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("upgrade", "--image", helloworldImage, name, "/hello")
	c.Assert(res.Error, check.IsNil)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output: %s, expected: %s", out, name)
	}
}

// TestPouchUpgradeCheckVolume is to verify if inherit old container's volume
// after upgrade a container
func (suite *PouchUpgradeSuite) TestPouchUpgradeCheckVolume(c *check.C) {
	name := "TestPouchUpgradeCheckVolume"

	// create container with a /data volume
	command.PouchRun("run", "-d", "-v", "/data", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// create a file in volume and write some data to the file
	command.PouchRun("exec", name, "sh", "-c", "echo '5678' >> /data/test").Assert(c, icmd.Success)

	res := command.PouchRun("upgrade", "--image", busyboxImage125, name)
	res.Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output: %s, expected: %s", out, name)
	}

	// check if the new container is running after upgade a running container
	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].State.Running, check.Equals, true)

	// double check if container is running by executing a exec command
	out := command.PouchRun("exec", name, "cat", "/data/test").Stdout()
	if !strings.Contains(out, "5678") {
		c.Errorf("failed to exec in container, expected 5678 got %s", out)
	}
}
