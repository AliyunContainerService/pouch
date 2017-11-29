package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchCreateSuite is the test suite fo help CLI.
type PouchCreateSuite struct {
}

func init() {
	check.Suite(&PouchCreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchCreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchCreateName is to verify the correctness of creating contaier with specified name.
func (suite *PouchCreateSuite) TestPouchCreateName(c *check.C) {
	containerName := "foo"
	out, err := exec.Command("pouch", "create", "--name", containerName, "busybox:latest").Output()
	c.Assert(err, check.IsNil)

	if !strings.Contains(string(out), containerName) {
		c.Fatalf("unexpected output %s expected %s\n", string(out), containerName)
	}
}

// TestPouchCreateWithVolumes tests if binded volumes are located in container's config.
func (suite *PouchCreateSuite) TestPouchCreateWithVolumes(c *check.C) {
	// TODO: add test when inspect API and CLI are supported
}

// TestPouchCreateWithTTY test if tty works correctly.
func (suite *PouchCreateSuite) TestPouchCreateWithTTY(c *check.C) {
	// TODO: add test with TTY
}
