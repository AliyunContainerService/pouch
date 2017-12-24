package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIContainerExecSuite is the test suite for container exec API.
type APIContainerExecSuite struct{}

func init() {
	check.Suite(&APIContainerExecSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerExecSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestContainerExecOk tests execing containers is OK.
func (suite *APIContainerExecSuite) TestContainerExecOk(c *check.C) {
	// TODO:
	// path := "/containers/{name:.*}/exec"
}
