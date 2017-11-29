package main

import (
	"github.com/go-check/check"
)

// PouchVolumeRmSuite is the test suite fo volume rm CLI.
type PouchVolumeRmSuite struct {
}

func init() {
	check.Suite(&PouchVolumeRmSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchVolumeRmSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchVolumeRm is to verify the correctness of volume rm command.
func (suite *PouchVolumeRmSuite) TestPouchVolumeRm(c *check.C) {

}
