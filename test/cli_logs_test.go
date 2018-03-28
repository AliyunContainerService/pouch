package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// PouchLogsSuite is the test suite for logs CLI.
type PouchLogsSuite struct{}

func init() {
	check.Suite(&PouchLogsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchLogsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchLogsSuite) TeadDownTest(c *check.C) {
	// TODO
}
