package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIContainerResizeSuite is the test suite for container upgrade API.
type APIContainerResizeSuite struct{}

func init() {
	check.Suite(&APIContainerResizeSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerResizeSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

}

// TODO add test case.
