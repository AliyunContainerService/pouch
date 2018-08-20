package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIContainerListSuite is the test suite for container list API.
type APIContainerListSuite struct{}

func init() {
	check.Suite(&APIContainerListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestListOk test api is ok with default parameters.
func (suite *APIContainerListSuite) TestListOk(c *check.C) {
	// TODO: missing case
	helpwantedForMissingCase(c, "container api list 200 case")
}
