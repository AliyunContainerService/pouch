package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchStartSuite is the test suite fo start CLI.
type PouchStartSuite struct {
}

func init() {
	check.Suite(&PouchPullSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchStartSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchPull is to verify the correctness of pull command
func (suite *PouchStartSuite) TestPouchStart(c *check.C) {

}
