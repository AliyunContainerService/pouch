package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// PouchAPIImageSearchSuite is the test suite for image search API.
type PouchAPIImageSearchSuite struct{}

func init() {
	check.Suite(&PouchAPIImageSearchSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIImageSearchSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageSearchOk tests searching images is OK.
func (suite *PouchAPIImageSearchSuite) TestImageSearchOk(c *check.C) {
	// TODO
	// path := "/images/search"
}
