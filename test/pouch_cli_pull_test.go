package main

import (
	"github.com/go-check/check"
)

// PouchPullSuite is the test suite fo help CLI.
type PouchPullSuite struct {
}

func init() {
	check.Suite(&PouchPullSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchPullSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchPullSuite) SetUpTest(c *check.C) {
	// TODO
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchPullSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchPullSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestPullWorks tests "pouch pull" work.
func (suite *PouchPullSuite) TestPullWorks(c *check.C) {

	// TODO: add wrong args.
}
