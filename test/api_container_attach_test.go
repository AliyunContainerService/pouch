package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIContainerAttachSuite is the test suite for container attach API.
type APIContainerAttachSuite struct{}

func init() {
	check.Suite(&APIContainerAttachSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerAttachSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestContainerAttachStdin tests attaching stdin is OK.
func (suite *APIContainerAttachSuite) TestContainerAttachStdin(c *check.C) {
	// TODO: missing case
	helpwantedForMissingCase(c, "container api attach 200 case")
}

// TestContainerAttachNotFound
func (suite *APIContainerAttachSuite) TestContainerAttachNotFound(c *check.C) {
	// TODO: missing case
	helpwantedForMissingCase(c, "container api attach 404 case")
}
