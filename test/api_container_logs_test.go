package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIContainerLogsSuite is the test suite for container logs API.
type APIContainerLogsSuite struct{}

func init() {
	check.Suite(&APIContainerLogsSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerLogsSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}
