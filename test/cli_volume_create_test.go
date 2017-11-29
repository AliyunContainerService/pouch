package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchVolumeCreateSuite is the test suite fo volume create CLI.
type PouchVolumeCreateSuite struct {
}

func init() {
	check.Suite(&PouchVolumeCreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchVolumeCreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchVolumeCreate is to verify the correctness of volume create command.
func (suite *PouchVolumeCreateSuite) TestPouchVolumeCreate(c *check.C) {

}
