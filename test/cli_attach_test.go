package main

import (
	"io"
	"os/exec"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/stretchr/testify/assert"
)

// PouchCreateSuite is the test suite for attach CLI.
type PouchAttachSuite struct{}

func init() {
	check.Suite(&PouchAttachSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchAttachSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchAttachSuite) TearDownTest(c *check.C) {
}

// TestPouchAttachRunningContainer is to verify the correctness of attach a running container.
func (suite *PouchAttachSuite) TestPouchAttachRunningContainer(c *check.C) {
	name := "TestPouchAttachRunningContainer"

	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "/bin/sh", "-c", "while true; do echo hello; done")

	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	cmd := exec.Command(environment.PouchBinary, "attach", name)

	out, err := cmd.StdoutPipe()
	if err != nil {
		c.Fatal(err)
	}
	defer out.Close()

	if err := cmd.Start(); err != nil {
		c.Fatal(err)
	}

	buf := make([]byte, 1024)

	if _, err := out.Read(buf); err != nil && err != io.EOF {
		c.Fatal(err)
	}

	if !strings.Contains(string(buf), "hello") {
		c.Fatalf("unexpected output %s expected hello\n", string(buf))
	}
}

// TestAttachWithTty tests running container with -tty flag and attach stdin in a non-tty client.
func (suite *PouchAttachSuite) TestAttachWithTty(c *check.C) {
	name := "TestAttachWithTty"
	command.PouchRun("run", "-d", "-t", "--name", name, busyboxImage, "sleep", "100000").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)
	attachRes := command.PouchRun("attach", name)
	errString := attachRes.Stderr()
	assert.Equal(c, errString, "Error: the input device is not a TTY\n")
}
