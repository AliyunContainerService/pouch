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

	// TODO: missing case
	helpwantedForMissingCase(c, "image api search cases")
}
