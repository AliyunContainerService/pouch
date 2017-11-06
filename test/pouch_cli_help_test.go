package main

import (
	"os/exec"

	"github.com/go-check/check"
)

// PouchHelpSuite is the test suite fo help CLI.
type PouchHelpSuite struct {
}

func init() {
	check.Suite(&PouchHelpSuite{})
}

// SetUpTest does common setup in the begining of each test.
func (suite *PouchHelpSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestExample is a demo of CLI test.
func (suite *PouchHelpSuite) TestExample(c *check.C) {
	cmd := exec.Command("ls")
	err := cmd.Run()
	c.Assert(err, check.IsNil)
}
