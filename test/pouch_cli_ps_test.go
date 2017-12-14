package main

import (
	"os/exec"

	"github.com/go-check/check"
)

// PouchPsSuite is the test suite fo help CLI.
type PouchPsSuite struct {
}

func init() {
	check.Suite(&PouchPsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchPsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

	// Pull test image
	cmd := exec.Command("pouch", "pull", testImage)
	cmd.Run()
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchPsSuite) SetUpTest(c *check.C) {
	// TODO
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchPsSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchPsSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestPsWorks tests "pouch ps" work.
func (suite *PouchPsSuite) TestPsWorks(c *check.C) {

	// TODO: add wrong args.
	// TODO: nil input should return success
	args := map[string]bool{
		"": false,
	}

	for arg, ok := range args {
		cmd := PouchCmd{
			args:   []string{"ps", arg},
			result: ok,
		}
		RunCmd(c, &cmd)
	}
}
