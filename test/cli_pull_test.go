package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchPullSuite is the test suite fo ps CLI.
type PouchPullSuite struct {
}

func init() {
	check.Suite(&PouchPullSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchPullSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchPull is to verify the correctness of pull command
func (suite *PouchPullSuite) TestPouchPull(c *check.C) {

}
