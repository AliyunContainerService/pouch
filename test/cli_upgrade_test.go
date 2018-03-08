package main

import (
	//"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchUpgradeSuite is the test suite for upgrade CLI.
type PouchUpgradeSuite struct{}

func init() {
	check.Suite(&PouchUpgradeSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchUpgradeSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchUpgradeSuite) TeadDownTest(c *check.C) {
	// TODO
}
