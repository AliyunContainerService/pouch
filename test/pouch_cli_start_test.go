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
	containername := "foo2"
	cmd := PouchCmd{
		args:   []string{"create", "--name", containername, testImage},
		result: true,
	}
	RunCmd(c, &cmd)

	cmd = PouchCmd{
		args:   []string{"start", containername},
		result: true,
	}
	RunCmd(c, &cmd)
}
