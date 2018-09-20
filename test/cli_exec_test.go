package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
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

// TestExecWithEnvs is to verify Exec with Envs.
func (suite *PouchExecSuite) TestExecWithEnvs(c *check.C) {
	name := "exec-normal3"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "sleep", "100000")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	res = command.PouchRun("exec", "-e \"Test=OK\"", name, "env")

	if out := res.Combined(); !strings.Contains(out, "Test=OK") {
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
	name := "TestExecInteractiveContainer"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	// use pipe to act interactive
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	defer stdoutW.Close()

	// need to start command before write
	cmd := command.PouchCmd("exec", "-i", name, "cat")
	cmd.Stdin = stdinR
	cmd.Stdout = stdoutW
	res = icmd.StartCmd(cmd)

	// send bye to cat
	//
	// FIXME(fuweid): when we remove ringbuffer, we can remove the
	// stdoutW, stdoutR pipe.
	fmt.Fprintf(stdinW, "bye\n")
	got, _, err := bufio.NewReader(stdoutR).ReadLine()
	c.Assert(err, check.IsNil)
	c.Assert(string(got), check.Equals, "bye")

	// send the EOF to stdin
	//
	// FIXME(fuweid): remove the timeout when we remove ringbuffer
	stdinW.Close()
	res = icmd.WaitOnCmd(10*time.Second, res)
	res.Assert(c, icmd.Success)

	// check process has gone
	{
		res := command.PouchRun("exec", name, "ps")
		res.Assert(c, icmd.Success)
		if strings.Contains(res.Combined(), "cat") {
			c.Errorf("cat process should be gone, but got \n%s\n", res.Combined())
		}
	}
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

// TestExecFail test exec fail should not hang, and test failed exec exit code should not be zero.
func (suite *PouchExecSuite) TestExecFail(c *check.C) {
	name := "TestExecFail"
	res := command.PouchRun("run", "-d", "--name", name, "-u", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	c.Assert(res.Stderr(), check.NotNil)

	// test a 'executable file not found' fail should get exit code 126.
	name = "TestExecFailCode"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	command.PouchRun("exec", name, "shouldnotexit").Assert(c, icmd.Expected{ExitCode: 126})

	// test a 'ls /nosuchfile' fail should get exit code not equal to 0.
	res = command.PouchRun("exec", name, "ls", "/nosuchfile")
	if res.ExitCode == 0 {
		c.Fatalf("failed exec process exit code should not be 0")
	}
}

// TestExecUser test exec with user.
func (suite *PouchExecSuite) TestExecUser(c *check.C) {
	name := "TestExecUser"
	res := command.PouchRun("run", "-d", "-u=1001", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("exec", name, "id", "-u")
	res.Assert(c, icmd.Success)
	if !strings.Contains(res.Stdout(), "1001") {
		c.Fatalf("failed to run a container with expected user: %s, but got %s", "1001", res.Stdout())
	}

	res = command.PouchRun("exec", "-u=1002", name, "id", "-u")
	res.Assert(c, icmd.Success)
	if !strings.Contains(res.Stdout(), "1002") {
		c.Fatalf("failed to run a container with expected user: %s, but got %s", "1002", res.Stdout())
	}

	// test user should not changed by exec process
	res = command.PouchRun("exec", name, "id", "-u")
	res.Assert(c, icmd.Success)
	if !strings.Contains(res.Stdout(), "1001") {
		c.Fatalf("failed to run a container with expected user: %s, but got %s", "1001", res.Stdout())
	}
}

// TestExecWithWorkingDir test working directory of exec process
func (suite *PouchExecSuite) TestExecWithWorkingDir(c *check.C) {
	cname := "TestExecWithWorkingDir"

	dir := "/tmp/testworkingdir"

	res := command.PouchRun("run", "-d", "--name", cname, "-w", dir, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("exec", cname, "pwd")
	res.Assert(c, icmd.Success)
	if !strings.Contains(res.Stdout(), dir) {
		c.Fatalf("failed to run exec with specified working directory: %s, but got %s", dir, res.Stdout())
	}
}
