package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIContainerUpgradeSuite is the test suite for container upgrade API.
type APIContainerUpgradeSuite struct{}

func init() {
	check.Suite(&APIContainerUpgradeSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerUpgradeSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	// TODO: missing case
	helpwantedForMissingCase(c, "container api upgrade cases")
}
