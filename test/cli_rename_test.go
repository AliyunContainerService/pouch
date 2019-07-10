package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRenameSuite is the test suite for rename CLI.
type PouchRenameSuite struct{}

func init() {
	check.Suite(&PouchRenameSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRenameSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

func (suite *PouchRenameSuite) TestRenameInvalidName(c *check.C) {
	name := "myName"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("rename", "myName", "new:invalid")
	if !strings.Contains(res.Stdout(), "Invalid container name") {
		check.Commentf("Expected '%s', but got %q", "Invalid container name", res.Stdout())
	}
}
