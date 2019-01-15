package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/stretchr/testify/assert"
)

// PouchStartSuite is the test suite for start CLI.
type PouchStartSuite struct{}

func init() {
	check.Suite(&PouchStartSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchStartSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchStartSuite) TearDownTest(c *check.C) {
}

// TestStartCommand tests "pouch start" work.
func (suite *PouchStartSuite) TestStartCommand(c *check.C) {
	name := "start-normal"
	res := command.PouchRun("create", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
}

// TestStartWithStdin tests "pouch start -i" work.
func (suite *PouchStartSuite) TestStartWithStdin(c *check.C) {
	// make echo server
	name := c.TestName()
	command.PouchRun("create", "--name", name, "-i", busyboxImage, "cat").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// start start
	cmd := exec.Command(environment.PouchBinary, "start", "-a", "-i", name)

	// prepare the io stream
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

	msg := "Hello, Pouch"
	_, err = stdin.Write([]byte(msg + "\n"))
	c.Assert(err, check.IsNil)

	out, err := bufio.NewReader(stdout).ReadString('\n')
	c.Assert(err, check.IsNil)
	c.Assert(strings.TrimSpace(out), check.Equals, msg)
	c.Assert(stdin.Close(), check.IsNil)
}

// TestStartInWrongWay runs start command in wrong way.
func (suite *PouchStartSuite) TestStartInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "missing container name", args: ""},
		{name: "unknown flag", args: "-k"},
	} {
		res := command.PouchRun("start", tc.args)
		c.Assert(res.Stderr(), check.NotNil, check.Commentf(tc.name))
	}
}

// TestStartWithEnv starts a container with env.
func (suite *PouchStartSuite) TestStartWithEnv(c *check.C) {
	name := "start-env"
	env := "abc=123"

	res := command.PouchRun("create", "--name", name, "-e", env, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("start", name).Assert(c, icmd.Success)
	output := command.PouchRun("exec", name, "/bin/env").Stdout()
	if !strings.Contains(output, env) {
		c.Errorf("failed to set env: %s, %s", env, output)
	}

	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
}

// TestStartWithEntrypoint starts a container with  entrypoint.
func (suite *PouchStartSuite) TestStartWithEntrypoint(c *check.C) {
	name := "start-entrypoint"

	command.PouchRun("create", "--name", name, "--entrypoint", "sh", busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	//TODO: check entrypoint really works
}

// TestStartWithWorkDir starts a container with work dir.
func (suite *PouchStartSuite) TestStartWithWorkDir(c *check.C) {
	name := "start-workdir"

	command.PouchRun("create", "--name", name, "--entrypoint", "pwd",
		"-w", "/tmp", busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("start", "-a", name).Stdout()
	if !strings.Contains(output, "/tmp") {
		c.Errorf("failed to start a container with workdir: %s", output)
	}
}

// TestStartWithUser starts a container with user.
func (suite *PouchStartSuite) TestStartWithUser(c *check.C) {
	name := "start-user"
	user := "1001"
	group := "1001"

	command.PouchRun("create", "--name", name, "--user", user, busyboxImage, "id", "-u")
	defer DelContainerForceMultyTime(c, name)
	output := command.PouchRun("start", "-a", name).Stdout()
	if !strings.Contains(output, user) {
		c.Errorf("failed to start a container with user: %s", output)
	}

	name = "start-group"
	command.PouchRun("create", "--name", name, "--user", user+":"+group, busyboxImage, "id", "-g")
	defer DelContainerForceMultyTime(c, name)

	output = command.PouchRun("start", "-a", name).Stdout()
	if !strings.Contains(output, group) {
		c.Errorf("failed to start a container with user:group : %s", output)
	}
}

// TestStartWithHostname starts a container with hostname.
func (suite *PouchStartSuite) TestStartWithHostname(c *check.C) {
	name := "start-hostname"
	hostname := "pouch"

	command.PouchRun("create", "--name", name, "--hostname", hostname, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
	output := command.PouchRun("exec", name, "hostname").Stdout()
	if !strings.Contains(output, hostname) {
		c.Errorf("failed to set hostname: %s, %s", hostname, output)
	}

	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
}

// TestStartWithSysctls starts a container with sysctls.
func (suite *PouchStartSuite) TestStartWithSysctls(c *check.C) {
	sysctl := "net.ipv4.ip_forward=1"
	name := "start-sysctl"

	command.PouchRun("create", "--name", name, "--sysctl", sysctl, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
	output := command.PouchRun("exec", name, "cat", "/proc/sys/net/ipv4/ip_forward").Stdout()
	if !strings.Contains(output, "1") {
		c.Errorf("failed to start a container with sysctls: %s", output)
	}

	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
}

// TestStartWithAppArmor starts a container with security option AppArmor.
func (suite *PouchStartSuite) TestStartWithAppArmor(c *check.C) {
	appArmor := "apparmor=unconfined"
	name := "start-apparmor"

	command.PouchRun("create", "--name", name, "--security-opt", appArmor, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	// TODO: do the test more strictly with effective AppArmor profile.

	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
}

// TestStartWithSeccomp starts a container with security option seccomp.
func (suite *PouchStartSuite) TestStartWithSeccomp(c *check.C) {
	seccomp := "seccomp=unconfined"
	name := "start-seccomp"

	command.PouchRun("create", "--name", name, "--security-opt", seccomp, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	// TODO: do the test more strictly with effective seccomp profile.

	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
}

// TestStartWithCapability starts a container with capability.
func (suite *PouchStartSuite) TestStartWithCapability(c *check.C) {
	capability := "NET_ADMIN"
	name := "start-capability"

	res := command.PouchRun("create", "--name", name, "--cap-add", capability, busyboxImage, "brctl", "addbr", "foobar")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
}

// TestStartWithPrivilege starts a container with privilege.
func (suite *PouchStartSuite) TestStartWithPrivilege(c *check.C) {
	name := "start-privilege"

	res := command.PouchRun("create", "--name", name, "--privileged", busyboxImage, "brctl", "addbr", "foobar")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
}

// TestStartWithAnnotation starts a container with annotation.
func (suite *PouchStartSuite) TestStartWithAnnotation(c *check.C) {
	name := "start-annotation"

	res := command.PouchRun("create", "--name", name, "--annotation", "a=b", busyboxImage, "top")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
}

// TestStartWithExitCode starts a container with annotation.
func (suite *PouchStartSuite) TestStartWithExitCode(c *check.C) {
	name := "start-exitcode"

	res := command.PouchRun("create", "--name", name, busyboxImage, "sh", "-c", "exit 101")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// test process exit code $? == 101
	ret := command.PouchRun("start", "-a", name)
	ret.Assert(c, icmd.Expected{ExitCode: 101})

	// test container ExitCode == 101
	exitCode, err := inspectFilter(name, ".State.ExitCode")
	c.Assert(err, check.IsNil)
	c.Assert(exitCode, check.Equals, "101")
}

// TestStartWithUlimit starts a container with --ulimit.
func (suite *PouchStartSuite) TestStartWithUlimit(c *check.C) {
	name := "start-ulimit"

	res := command.PouchRun("create", "--name", name,
		"--ulimit", "nproc=256", busyboxImage, "top")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
}

// TestStartWithPidsLimit tests running container with --pids-limit flag.
func (suite *PouchStartSuite) TestStartWithPidsLimit(c *check.C) {
	name := "TestStartWithPidsLimit"
	pidfile := "/sys/fs/cgroup/pids/pids.max"
	res := command.PouchRun("create", "--pids-limit", "10",
		"--name", name, busyboxImage, "cat", pidfile)
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
}

// TestStartFromCheckpoint tests start a container from a checkpoint
func (suite *PouchStartSuite) TestStartFromCheckpoint(c *check.C) {
	SkipIfFalse(c, environment.IsCRIUExist)
	name := "TestStartFromCheckpoint"
	defer DelContainerForceMultyTime(c, name)
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

	tmpDir, err := ioutil.TempDir("", "checkpoint")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(tmpDir)
	checkpoint := "cp0"
	command.PouchRun("checkpoint", "create", "--checkpoint-dir", tmpDir, name, checkpoint).Assert(c, icmd.Success)
	// check criu image files have been dumped into checkpoint-dir, pouch create a description json,
	// so there should be more than 1 files
	dirs, err := ioutil.ReadDir(filepath.Join(tmpDir, checkpoint))
	c.Assert(err, check.IsNil)
	if len(dirs) < 2 {
		c.Errorf("failed to dump criu image for container %s", name)
	}

	restoredContainer := "restoredContainer"
	defer DelContainerForceMultyTime(c, restoredContainer)
	// image busybox not have /proc directory, we need to start busybox image and stop it
	// make /proc exist, then we can restore successful
	command.PouchRun("run", "-d", "--name", restoredContainer, busyboxImage).Assert(c, icmd.Success)
	command.PouchRun("stop", "-t", "1", restoredContainer).Assert(c, icmd.Success)

	command.PouchRun("start", "--checkpoint-dir", tmpDir, "--checkpoint", checkpoint, restoredContainer).Assert(c, icmd.Success)

	result := command.PouchRun("exec", restoredContainer, "sh", "-c", "ps -ef | grep top").Assert(c, icmd.Success)
	if !strings.Contains(result.Stdout(), "top") {
		c.Error("restored container should have top process")
	}
}

// TestStartWithTty tests running container with -tty flag and attach stdin in a non-tty client.
func (suite *PouchStartSuite) TestStartWithTty(c *check.C) {
	name := "TestStartWithTty"
	res := command.PouchRun("create", "-t", "--name", name, busyboxImage, "/bin/sh", "-c", "while true;do echo hello;done")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	attachRes := command.PouchRun("start", "-a", "-i", name)
	errString := attachRes.Stderr()
	assert.Equal(c, errString, "Error: the input device is not a TTY\n")
}

// TestStartMultiContainers tries to start more than one container.
func (suite *PouchStartSuite) TestStartMultiContainers(c *check.C) {
	containernames := []string{"TestStartMultiContainer-1", "TestStartMultiContainer-2"}
	for _, name := range containernames {
		res := command.PouchRun("create", "--name", name, busyboxImage, "top")
		defer DelContainerForceMultyTime(c, name)
		res.Assert(c, icmd.Success)
	}

	res := command.PouchRun("start", containernames[0], containernames[1])
	res.Assert(c, icmd.Success)

	res = command.PouchRun("stop", "-t", "1", containernames[0], containernames[1])
	res.Assert(c, icmd.Success)
}

// TestStartContainerTwice tries to start a container twice
func (suite *PouchStartSuite) TestStartContainerTwice(c *check.C) {
	name := "TestStartContainerTwice"
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("create", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	command.PouchRun("start", name).Assert(c, icmd.Success)
	command.PouchRun("start", name).Assert(c, icmd.Success)
}
