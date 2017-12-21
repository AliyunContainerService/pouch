package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// PouchAPIContainerExecSuite is the test suite for container exec API.
type PouchAPIContainerExecSuite struct{}

func init() {
	check.Suite(&PouchAPIContainerExecSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIContainerExecSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestContainerExecOk tests execing containers is OK.
func (suite *PouchAPIContainerExecSuite) TestContainerExecOk(c *check.C) {
	// TODO
	// path := "/containers/{name:.*}/exec"
}
