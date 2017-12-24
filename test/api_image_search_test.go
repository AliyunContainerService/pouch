package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIImageSearchSuite is the test suite for image search API.
type APIImageSearchSuite struct{}

func init() {
	check.Suite(&APIImageSearchSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageSearchSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageSearchOk tests searching images is OK.
func (suite *APIImageSearchSuite) TestImageSearchOk(c *check.C) {
	// TODO
	// path := "/images/search"
}
