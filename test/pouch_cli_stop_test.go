package main

import (
	"os/exec"

	"github.com/go-check/check"
)

// PouchStopSuite is the test suite fo help CLI.
type PouchStopSuite struct {
}

func init() {
	check.Suite(&PouchStopSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchStopSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

	// Pull test image
	cmd := exec.Command("pouch", "pull", testImage)
	cmd.Run()
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchStopSuite) SetUpTest(c *check.C) {
	// TODO
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchStopSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchStopSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestStopWorks tests "pouch stop" work.
func (suite *PouchStopSuite) TestStopWorks(c *check.C) {

}
