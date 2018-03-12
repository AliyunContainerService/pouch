package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchTopSuite is the test suite for top CLI.
type PouchTopSuite struct{}

func init() {
	check.Suite(&PouchTopSuite{})
}

// SetupSuite does common setup in the beginning of each test suite.
func (suite *PouchTopSuite) SetupSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchTopSuite) TearDownTest(c *check.C) {
}
