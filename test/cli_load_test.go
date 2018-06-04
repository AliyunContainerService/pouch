package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// PouchLoadSuite is the test suite for load CLI.
type PouchLoadSuite struct{}

func init() {
	check.Suite(&PouchLoadSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchLoadSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TODO(fuwei): We cannot upload the oci.v1 format tar into repo because it will
// increase our repo size. The test will be done with "pouch save" functionality.
