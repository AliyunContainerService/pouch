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

// TestHelpWorks is a demo of CLI test.
func (suite *PouchHelpSuite) TestHelpWorks(c *check.C) {
	// TODO: add wrong args.
	args := map[string]bool{
		"help":   true,
		"--help": true,
		"-h":     true,
	}

	for arg, ok := range args {
		cmd := exec.Command("pouch", arg)
		_, _, err := runCmd(cmd)

		if ok {
			c.Assert(err, check.IsNil)
		} else {
			c.Assert(err, check.NotNil)
		}
	}
}
