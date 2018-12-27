package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/stretchr/testify/assert"
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

func (suite *PouchExecSuite) TestExecNoCommand(c *check.C) {
	cname := "TestExecNoCommand"
	command.PouchRun("run", "-d", "--name", cname, busyboxImage, "sleep", "100000").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	res := command.PouchRun("exec", cname)
	expectedError := "requires at least 2 arg(s), only received 1"
	if out := res.Combined(); !strings.Contains(out, expectedError) {
		c.Fatalf("unexpected output %s, expected %s", out, expectedError)
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

// TestExecWithReplaceEnvs is to verify New Envs will replace old Envs.
func (suite *PouchExecSuite) TestExecWithReplaceEnvs(c *check.C) {
	name := "exec-normal4"
	res := command.PouchRun("run", "-d", "-e", "Test=Old", "--name", name, busyboxImage, "sleep", "100000")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	res = command.PouchRun("exec", name, "env")

	if out := res.Combined(); !strings.Contains(out, "Test=Old") {
		c.Fatalf("unexpected output %s expected %s\n", out, name)
	}

	res = command.PouchRun("exec", "-e \"Test=New\"", name, "env")

	if out := res.Combined(); !strings.Contains(out, "Test=New") {
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

	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)

	out := command.PouchRun("exec", name, "echo", "test").Stderr()
	if !strings.Contains(out, "failed") {
		c.Errorf("should fail to exec in stopped container: %s", out)
	}
}

// TestExecInteractive test "-i" option works.
func (suite *PouchExecSuite) TestExecInteractive(c *check.C) {
	name := "TestExecInteractive"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "sleep", "100000").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	cmd := exec.Command(environment.PouchBinary, "exec", "-i", name, "sh")
	stdin, err := cmd.StdinPipe()
	c.Assert(err, check.IsNil)
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	c.Assert(err, check.IsNil)
	defer stdout.Close()
	c.Assert(cmd.Start(), check.IsNil)
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	_, err = stdin.Write([]byte("echo hello\n"))
	c.Assert(err, check.IsNil)
	out, err := bufio.NewReader(stdout).ReadString('\n')
	c.Assert(err, check.IsNil)
	c.Assert(strings.TrimSpace(out), check.Equals, "hello")

	c.Assert(stdin.Close(), check.IsNil)
}

// TestExecAfterContainerRestart test exec in a restart container should work.
func (suite *PouchExecSuite) TestExecAfterContainerRestart(c *check.C) {
	name := "TestExecAfterContainerRestart"
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)

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
	defer DelContainerForceMultyTime(c, name)
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

// TestExecWithTty tests running container with -tty flag and attach stdin in a non-tty client.
func (suite *PouchExecSuite) TestExecWithTty(c *check.C) {
	name := "TestExecWithTty"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "sleep", "100000").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)
	attachRes := command.PouchRun("exec", "-i", "-t", name, "ls")
	errString := attachRes.Stderr()
	assert.Equal(c, errString, "Error: the input device is not a TTY\n")
}

// TestExecForCloseIO test CloseIO works.
func (suite *PouchExecSuite) TestExecForCloseIO(c *check.C) {
	name := "TestExecForCloseIO"
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cmdLine := fmt.Sprintf("echo 1 | %s exec -i %s sh -c 'cat && echo hello'", environment.PouchBinary, name)
	out, err := exec.CommandContext(ctx, "bash", "-c", cmdLine).Output()
	c.Assert(err, check.IsNil)
	c.Assert(string(out), check.Equals, "1\nhello\n")
}

// TestExecWithPrivileged tests exec with --privileged can work
func (suite *PouchExecSuite) TestExecWithPrivileged(c *check.C) {
	name := "TestExecWithPrivileged"
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("run", "-d", "--name", name, "--cap-drop=ALL", busyboxImage, "top").Assert(c, icmd.Success)

	// without --privileged, exec should fails
	command.PouchRun("exec", name, "sh", "-c", "mknod /tmp/sda b 8 16").Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "Operation not permitted",
	})

	command.PouchRun("exec", "--privileged", name, "sh", "-c", "mknod /tmp/sdb b 8 16").Assert(c, icmd.Success)
	ret := command.PouchRun("exec", name, "ls", "/tmp/sdb")
	ret.Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "/tmp/sdb\n")
}
