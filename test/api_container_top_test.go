package main

import (
	"github.com/alibaba/pouch/test/environment"
	//"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerTopSuite is the test suite for container top API.
type APIContainerTopSuite struct{}

func init() {
	check.Suite(&APIContainerTopSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerTopSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	// TODO Add more
}
