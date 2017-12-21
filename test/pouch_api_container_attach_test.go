package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// PouchAPIContainerAttachSuite is the test suite for container attach API.
type PouchAPIContainerAttachSuite struct{}

func init() {
	check.Suite(&PouchAPIContainerAttachSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIContainerAttachSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestContainerAttachOk tests attaching containers is OK.
func (suite *PouchAPIContainerAttachSuite) TestContainerAttachOk(c *check.C) {
	// TODO
	// path := "/containers/{name:.*}/attach"
}
