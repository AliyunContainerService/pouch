package main

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/kr/pty"
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
	command.PouchRun("create", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	command.PouchRun("stop", name).Assert(c, icmd.Success)

	defer DelContainerForceMultyTime(c, name)
}

// TestStartInTTY tests "pouch start -i" work.
func (suite *PouchStartSuite) TestStartInTTY(c *check.C) {
	// make echo server
	name := "start-tty"
	command.PouchRun("create", "--name", name, busyboxImage, "cat").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// start tty and redirect
	cmd := exec.Command(environment.PouchBinary, "start", "-a", "-i", name)
	fd, err := pty.Start(cmd)
	c.Assert(err, check.IsNil)
	defer fd.Close()

	msg := "Hello, Pouch"

	// hey
	_, err = fd.Write([]byte(msg + "\n"))
	c.Assert(err, check.IsNil)

	// what?
	echo, err := bufio.NewReader(fd).ReadString('\n')
	c.Assert(err, check.IsNil)
	c.Assert(strings.TrimSpace(echo), check.Equals, msg)

	command.PouchRun("stop", name)
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
		c.Assert(res.Error, check.NotNil, check.Commentf(tc.name))
	}
}

// TestStartWithEnv starts a container with env.
func (suite *PouchStartSuite) TestStartWithEnv(c *check.C) {
	name := "start-env"
	env := "abc=123"

	command.PouchRun("create", "--name", name, "-e", env, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
	output := command.PouchRun("exec", name, "/bin/env").Stdout()
	if !strings.Contains(output, env) {
		c.Errorf("failed to set env: %s, %s", env, output)
	}

	command.PouchRun("stop", name).Assert(c, icmd.Success)
}

// TestStartWithEntrypoint starts a container with  entrypoint.
func (suite *PouchStartSuite) TestStartWithEntrypoint(c *check.C) {
	name := "start-entrypoint"

	command.PouchRun("create", "--name", name, "--entrypoint", "sh", busyboxImage).Assert(c, icmd.Success)
	command.PouchRun("start", name).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	//TODO: check entrypoint really works
}

// TestStartWithWorkDir starts a container with work dir.
func (suite *PouchStartSuite) TestStartWithWorkDir(c *check.C) {
	name := "start-workdir"

	command.PouchRun("create", "--name", name, "--entrypoint", "pwd", "-w", "/tmp", busyboxImage).Assert(c, icmd.Success)
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
	output := command.PouchRun("start", "-a", name).Stdout()
	if !strings.Contains(output, user) {
		c.Errorf("failed to start a container with user: %s", output)
	}

	name = "start-group"
	command.PouchRun("create", "--name", name, "--user", user+":"+group, busyboxImage, "id", "-g")
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

	command.PouchRun("stop", name).Assert(c, icmd.Success)
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

	command.PouchRun("stop", name).Assert(c, icmd.Success)
}

// TestStartWithAppArmor starts a container with security option AppArmor.
func (suite *PouchStartSuite) TestStartWithAppArmor(c *check.C) {
	appArmor := "apparmor=unconfined"
	name := "start-apparmor"

	command.PouchRun("create", "--name", name, "--security-opt", appArmor, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	// TODO: do the test more strictly with effective AppArmor profile.

	command.PouchRun("stop", name).Assert(c, icmd.Success)
}

// TestStartWithSeccomp starts a container with security option seccomp.
func (suite *PouchStartSuite) TestStartWithSeccomp(c *check.C) {
	seccomp := "seccomp=unconfined"
	name := "start-seccomp"

	command.PouchRun("create", "--name", name, "--security-opt", seccomp, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	// TODO: do the test more strictly with effective seccomp profile.

	command.PouchRun("stop", name).Assert(c, icmd.Success)
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

	res := command.PouchRun("create", "--name", name, "--annotation", "a=b", busyboxImage)
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
	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].State.ExitCode, check.Equals, int64(101))
}

// TestStartWithUlimit starts a container with --ulimit.
func (suite *PouchStartSuite) TestStartWithUlimit(c *check.C) {
	name := "start-ulimit"

	res := command.PouchRun("create", "--name", name, "--ulimit", "nproc=256", busyboxImage)
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)
}
