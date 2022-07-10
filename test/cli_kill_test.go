package main

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchKillSuite is the test suite for kill CLI.
type PouchKillSuite struct{}

func init() {
	check.Suite(&PouchKillSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchKillSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

func (suite *PouchKillSuite) TestKillContainer(c *check.C) {
	name := "TestKillContainer"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("inspect", "--format", `"{{.State.Status}}"`, name)
	if !strings.Contains(res.Stdout(), "running") {
		check.Commentf("Container %s is not running", name)
	}

	res = command.PouchRun("kill", name)
	if !strings.Contains(res.Stdout(), name) {
		check.Commentf("killed container is still running")
	}
}

func (suite *PouchKillSuite) TestKillOffStoppedContainer(c *check.C) {
	name := "TestKillOffStoppedContainer"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)
	command.PouchRun("stop", name).Assert(c, icmd.Success)

	res := command.PouchRun("kill", "-s", "30", name)
	err := fmt.Sprintf("Container %s is not running", name)
	if !strings.Contains(res.Stdout(), err) {
		check.Commentf(err)
	}
}

// regression test about correct signal parsing
func (suite *PouchKillSuite) TestKillWithSignal(c *check.C) {
	name := "TestKillWithSignal"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("kill", "-s", "SIGWINCH", name).Assert(c, icmd.Success)
	res := command.PouchRun("inspect", "--format", `"{{.State.Running}}"`, name)
	if !strings.Contains(res.Stdout(), "true") {
		check.Commentf("Container should be in running state after SIGWINCH")
	}
}

func (suite *PouchKillSuite) TestKillWithInvalidSignal(c *check.C) {
	name := "TestKillWithInvalidSignal"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("kill", "-s", "0")
	if !strings.Contains(res.Stdout(), "Invalid signal: 0") {
		check.Commentf("Kill with an invalid signal didn't error out correctly")
	}
	res = command.PouchRun("inspect", "--format", `"{{.State.Running}}"`, name)
	if !strings.Contains(res.Stdout(), "true") {
		check.Commentf("Container should be in running state after an invalid signal")
	}
}
