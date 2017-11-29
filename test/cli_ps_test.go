package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchPsSuite is the test suite fo ps CLI.
type PouchPsSuite struct {
}

func init() {
	check.Suite(&PouchPsSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchPsSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchPs is to verify the correctness of ps command
func (suite *PouchPsSuite) TestPouchPs(c *check.C) {

}
