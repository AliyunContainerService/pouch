package main

import (
	"github.com/go-check/check"
)

// PouchHelpSuite is the test suite fo help CLI.
type PouchHelpSuite struct {
}

func init() {
	check.Suite(&PouchHelpSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchHelpSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestHelpWorks tests "pouch help" work.
func (suite *PouchHelpSuite) TestHelpWorks(c *check.C) {
	// TODO: add wrong args.
	args := map[string]bool{
		"help":   true,
		"--help": true,
		"-h":     true,
		"-help":  false,
		"--h":    false,
	}

	for arg, ok := range args {
		cmd := PouchCmd{
			args:   []string{arg},
			result: ok,
		}
		RunCmd(c, &cmd)
	}
}
