package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchSuite is the test suite for pouch cli.
type PouchSuite struct{}

func init() {
	check.Suite(&PouchSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageTagOKWithSourceImageName tests OK.
func (suite *PouchSuite) TestInterspersedFlags(c *check.C) {
	// set flags directly to pouch cli
	command.PouchRun("-D", "ps").Assert(c, icmd.Success)

	// set pouch cli flags to subcommand is forbidden
	command.PouchRun("pull",
		"--tlscacert",
		"/data/cert/ca.crt",
		"--tlscert", "/data/cert/20.26.38.138.cert",
		"--tlskey", "/data/cert/20.26.38.138.key",
		busyboxImage,
	).Assert(c, icmd.Expected{
		ExitCode: 1, // non-zero exit code
		Error:    "Error: unknown flag: --tlscacert",
	})
}
