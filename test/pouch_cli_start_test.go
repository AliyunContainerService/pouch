package main

import (
	"os/exec"

	"github.com/go-check/check"
)

// PouchStartSuite is the test suite fo help CLI.
type PouchStartSuite struct {
}

func init() {
	check.Suite(&PouchStartSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchStartSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

	// Pull test image
	cmd := exec.Command("pouch", "pull", testImage)
	cmd.Run()
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchStartSuite) SetUpTest(c *check.C) {
	// TODO
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchStartSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchStartSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestStartWorks tests "pouch start" work.
func (suite *PouchStartSuite) TestStartWorks(c *check.C) {

	cmd := exec.Command("pouch", "create", "--name", "foo2", testImage)
	runCmdPos(c, cmd)

	cmd = exec.Command("pouch", "start", "foo2")
	runCmdPos(c, cmd)
}
