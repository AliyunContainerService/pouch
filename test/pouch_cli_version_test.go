package main

import (
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

	var cmd PouchCmd

	cmd = PouchCmd{
		args:        "version",
		result:      true,
		outContains: "APIVersion",
	}
	RunCmd(c, &cmd)
}
