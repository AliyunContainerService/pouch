package main

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchExecSuite is the test suite for exec CLI.
type PouchExecSuite struct{}

func init() {
	check.Suite(&PouchExecSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchExecSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchExecSuite) TearDownTest(c *check.C) {
}

// TestExecCommand is to verify the correctness of execing container with specified command.
func (suite *PouchExecSuite) TestExecCommand(c *check.C) {
	name := "exec-normal"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "sleep", "100000")
	defer DelContainerForceMultyTime(c, name)

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
	defer DelContainerForceMultyTime(c, name)

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

// TestExecEcho tests exec prints the output.
func (suite *PouchExecSuite) TestExecEcho(c *check.C) {
	name := "TestExecEcho"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	out := command.PouchRun("exec", name, "echo", "test").Stdout()
	if !strings.Contains(out, "test") {
		c.Errorf("failed to exec in container: %s", out)
	}
}

// TestExecStoppedContainer test exec in a stopped container fail.
func (suite *PouchExecSuite) TestExecStoppedContainer(c *check.C) {
	name := "TestExecStoppedContainer"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("stop", name).Assert(c, icmd.Success)

	out := command.PouchRun("exec", name, "echo", "test").Stderr()
	if !strings.Contains(out, "failed") {
		c.Errorf("should fail to exec in stopped container: %s", out)
	}
}

// TestExecInteractive test "-i" option works.
func (suite *PouchExecSuite) TestExecInteractive(c *check.C) {
	//TODO
}

// TestExecAfterContainerRestart test exec in a restart container should work.
func (suite *PouchExecSuite) TestExecAfterContainerRestart(c *check.C) {
	name := "TestExecAfterContainerRestart"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("stop", name).Assert(c, icmd.Success)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	out := command.PouchRun("exec", name, "echo", "test").Stdout()
	if !strings.Contains(out, "test") {
		c.Errorf("failed to exec in container: %s", out)
	}
}

// TestExecUlimit test ulimit set container.
func (suite *PouchExecSuite) TestExecUlimit(c *check.C) {
	name := "TestExecUlimit"
	res := command.PouchRun("run", "-d", "--name", name, "--ulimit", "nproc=256",
		busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	out := command.PouchRun("exec", name, "sh", "-c", "ulimit -p").Stdout()
	c.Assert(out, check.Equals, "256\n")
}

// TestExecExitCode test exit code after exec process exit.
func (suite *PouchExecSuite) TestExecExitCode(c *check.C) {
	name := "TestExecExitCode"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("exec", name, "sh", "-c", "exit 101").Assert(c, icmd.Expected{ExitCode: 101})
	command.PouchRun("exec", name, "sh", "-c", "exit 0").Assert(c, icmd.Success)
}

// TestExecWithContainerdKilled test containerd get unexpected killed, and will restore
func (suite *PouchExecSuite) TestExecWithContainerdKilled(c *check.C) {
	name := "TestExecWithContainerdKilled"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	res.Assert(c, icmd.Success)

	// give time for containerd restore, more than 3 second.
	data, err := exec.Command("pidof", "containerd").CombinedOutput()
	c.Assert(err, check.IsNil)
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	c.Assert(err, check.IsNil)

	syscall.Kill(pid, syscall.SIGTERM)
	time.Sleep(5)

	output := command.PouchRun("exec", name, "echo", "1").Stdout()
	c.Assert(output, check.Equals, "1\n")
	command.PouchRun("stop", name).Assert(c, icmd.Success)
	command.PouchRun("rm", name).Assert(c, icmd.Success)
}
