package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchExecSuite is the test suite fo exec CLI.
type PouchExecSuite struct{}

func init() {
	check.Suite(&PouchExecSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchExecSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	c.Assert(environment.PruneAllContainers(apiClient), check.IsNil)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchExecSuite) TearDownTest(c *check.C) {
	c.Assert(environment.PruneAllContainers(apiClient), check.IsNil)
}

// TestExecCommand is to verify the correctness of execing container with specified command.
func (suite *PouchExecSuite) TestExecCommand(c *check.C) {
	name := "exec-normal"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "sleep", "100000")

	res.Assert(c, icmd.Success)

	res = command.PouchRun("exec", name, "ls")
	// the result should be like the following:
	// root@ubuntu:~/# pouch exec bf50a0 ls
	// bin
	// dev
	// etc
	// home
	// ...
	if out := res.Combined(); !strings.Contains(out, "etc") {
		c.Fatalf("unexpected output %s expected %s\n", out, name)
	}
}

// TestExecMultiCommands is to verify the correctness of execing container with specified commands.
func (suite *PouchExecSuite) TestExecMultiCommands(c *check.C) {
	name := "exec-normal2"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "sleep", "100000")

	res.Assert(c, icmd.Success)

	res = command.PouchRun("exec", name, "ls", "/etc")
	// the result should be like the following:
	// root@ubuntu:~/# pouch exec bf50a0 ls /etc
	// group
	// localtime
	// passwd
	// shadow
	if out := res.Combined(); !strings.Contains(out, "passwd") {
		c.Fatalf("unexpected output %s expected %s\n", out, name)
	}
}
