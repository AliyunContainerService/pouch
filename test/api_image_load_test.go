package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIImageLoadSuite is the test suite for image load API.
type APIImageLoadSuite struct{}

func init() {
	check.Suite(&APIImageLoadSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageLoadSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TODO(fuwei): We cannot upload the oci.v1 format tar into repo because it will
// increase our repo size. The test will be done with "pouch save" functionality.
