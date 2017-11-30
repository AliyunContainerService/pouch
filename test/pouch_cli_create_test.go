package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchCreateSuite is the test suite of help CLI.
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
	out, err := exec.Command("pouch", "create", "--name", "foo", "busybox:latest").Output()
	c.Assert(err, check.IsNil)

	if !strings.Contains(string(out), "foo") {
		c.Fatalf("unexpected output %s expected foo\n", string(out))
	}
}

// TestPouchCreateDuplicateContainerName is to verify duplicate container names.
func (suite *PouchCreateSuite) TestPouchCreateDuplicateContainerName(c *check.C) {
	containername := "duplicate"
	out, err := exec.Command("pouch", "create", "--name", containername, "busybox:latest").Output()
	c.Assert(err, check.IsNil)

	out, err = exec.Command("pouch", "create", "--name", containername, "busybox:latest").CombinedOutput()
	c.Assert(err, check.NotNil)

	if !strings.Contains(string(out), "already exist") {
		c.Fatalf("unexpected output %s expected already exist\n", string(out))
	}
}
