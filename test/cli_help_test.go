package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchHelpSuite is the test suite for help CLI.
type PouchHelpSuite struct{}

func init() {
	check.Suite(&PouchHelpSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchHelpSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestHelpWorks tests "pouch help" work.
func (suite *PouchHelpSuite) TestHelpWorks(c *check.C) {
	args := map[string]bool{
		"help":      true,
		"--help":    true,
		"-h":        true,
		"-help":     false,
		"--h":       false,
		"--unknown": false,
	}

	for arg, ok := range args {
		res := command.PouchRun(arg)
		if ok {
			res.Assert(c, icmd.Success)
		} else {
			c.Assert(res.Stderr(), check.NotNil)
		}
	}
}
