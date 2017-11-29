package main

import (
	"github.com/go-check/check"
	"os/exec"
	"strings"
)

// PouchVersionSuite is the test suite of help CLI.
type PouchVersionSuite struct {
}

func init() {
	check.Suite(&PouchVersionSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchVersionSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchVersion is to verify pouch version.
func (suite *PouchVersionSuite) TestPouchVersion(c *check.C) {
	out, err := exec.Command("pouch", "version").Output()
	c.Assert(err, check.IsNil)

	if !strings.Contains(string(out), "APIVersion") {
		c.Fatalf("unexpected output %s expected APIVersion\n", string(out))
	}

}
