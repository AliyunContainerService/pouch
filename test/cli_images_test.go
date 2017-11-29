package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchImagesSuite is the test suite fo images CLI.
type PouchImagesSuite struct {
}

func init() {
	check.Suite(&PouchImagesSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchImagesSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchImages is to verify the correctness of images command
func (suite *PouchImagesSuite) TestPouchImages(c *check.C) {

}
