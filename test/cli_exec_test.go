package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchCreateSuite is the test suite fo exec CLI.
type PouchExecSuite struct {
}

func init() {
	check.Suite(&PouchExecSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchExecSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchExec tests if a common exec would work.
func (suite *PouchExecSuite) TestPouchExec(c *check.C) {
	
}

// TestPouchExecWithDetach tests if a common exec with detach flag would work.
func (suite *PouchExecSuite) TestPouchExecWithDetach(c *check.C) {
	
}

// TestPouchExecWithTTY tests if a common exec with detach flag would work.
func (suite *PouchExecSuite) TestPouchExecWithTTY(c *check.C) {
	
}

// TestPouchExecWithInteractive tests if a common exec with detach flag would work.
func (suite *PouchCreateSuite) TestPouchExecWithInteractive(c *check.C) {
	
}


