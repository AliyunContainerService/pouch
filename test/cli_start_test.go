package main

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/kr/pty"
)

// PouchStartSuite is the test suite fo help CLI.
type PouchStartSuite struct{}

func init() {
	check.Suite(&PouchStartSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchStartSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	c.Assert(environment.PruneAllContainers(apiClient), check.IsNil)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchStartSuite) TearDownTest(c *check.C) {
	c.Assert(environment.PruneAllContainers(apiClient), check.IsNil)
}

// TestStartCommand tests "pouch start" work.
func (suite *PouchStartSuite) TestStartCommand(c *check.C) {
	name := "start-normal"
	command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	command.PouchRun("stop", name).Assert(c, icmd.Success)
}

// TestStartInTTY tests "pouch start -i" work.
func (suite *PouchStartSuite) TestStartInTTY(c *check.C) {
	// make echo server
	name := "start-tty"
	command.PouchRun("create", "--name", name, busyboxImage, "cat").Assert(c, icmd.Success)

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

	command.PouchRun("create", "--name", name, "-e", env, busyboxImage).Assert(c, icmd.Success)

	command.PouchRun("start", name).Assert(c, icmd.Success)
	output := command.PouchRun("exec", name, "/bin/env").Stdout()
	if !strings.Contains(output, env) {
		c.Errorf("failed to set env: %s, %s", env, output)
	}

	command.PouchRun("stop", name).Assert(c, icmd.Success)
}
