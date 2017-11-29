package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchStopSuite is the test suite fo start CLI.
type PouchStopSuite struct {
}

func init() {
	check.Suite(&PouchStopSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchStopSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchPull is to verify the correctness of stop command
func (suite *PouchStopSuite) TestPouchStop(c *check.C) {

}
