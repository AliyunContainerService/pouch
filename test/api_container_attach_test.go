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

// TestContainerAttachOk tests attaching containers is OK.
func (suite *APIContainerAttachSuite) TestContainerAttachOk(c *check.C) {
	// TODO
	// path := "/containers/{name:.*}/attach"
}
