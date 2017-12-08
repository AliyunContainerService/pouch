package main

import (
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

// PouchRmiSuite is the test suite fo help CLI.
type PouchRmiSuite struct {
}

func init() {
	check.Suite(&PouchRmiSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRmiSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

	// Pull test image
	cmd := exec.Command("pouch", "pull", testImage)
	cmd.Run()
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchRmiSuite) SetUpTest(c *check.C) {
	// TODO
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchRmiSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRmiSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestRmiWorks tests "pouch rmi" work.
func (suite *PouchRmiSuite) TestRmiWorks(c *check.C) {

	// TODO: add wrong args.

	out, err := exec.Command("pouch", "rmi", testImage).Output()
	c.Assert(err, check.IsNil)

	if !strings.Contains(string(out), testImage) {
		c.Fatalf("unexpected output %s expected %s\n", string(out), testImage)
	}
}
