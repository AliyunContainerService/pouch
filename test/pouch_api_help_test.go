package main

import (
	"github.com/go-check/check"
)

// PouchAPIHelpSuite is the test suite fo help related API.
type PouchAPIHelpSuite struct {
}

func init() {
	check.Suite(&PouchAPIHelpSuite{})
}

// SetUpTest does common setup in the begining of each test.
func (suite *PouchAPIHelpSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestExample is a demo of API test.
func (suite *PouchAPIHelpSuite) TestExample(c *check.C) {
}
