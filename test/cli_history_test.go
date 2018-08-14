package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// PouchHistorySuite is the test suite for history CLI.
type PouchHistorySuite struct{}

func init() {
	check.Suite(&PouchHistorySuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchHistorySuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestHistoryWorks tests "pouch history" work.
func (suite *PouchHistorySuite) TestHistoryWorks(c *check.C) {
	// TODO: We shouldn't compare dockerhub's image history with a fixed string, that's too unreadable.
	// So this test case will be done when pouch enables build functionality.
}
