package main

import (
	"os/exec"
	"strings"

	"github.com/alibaba/pouch/version"

	"github.com/go-check/check"
)

// PouchVersionSuite is the test suite fo help CLI.
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

	if !strings.Contains(string(out), version.Version) {
		c.Fatalf("unexpected output %s expected %s\n", string(out), version.Version)
	}

	if !strings.Contains(string(out), version.APIVersion) {
		c.Fatalf("unexpected output %s expected %s\n", string(out), version.APIVersion)
	}
}
