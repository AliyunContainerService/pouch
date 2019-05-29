package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"
	"github.com/gotestyourself/gotestyourself/icmd"

	"github.com/go-check/check"
)

// PouchSearchSuite is the test suite for search CLI.
type PouchSearchSuite struct{}

func init() {
	check.Suite(&PouchSearchSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchSearchSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestSearchWorks tests "pouch search" work.
func (suite *PouchSearchSuite) TestSearchWorks(c *check.C) {
	res := command.PouchRun("search", "nginx")
	res.Assert(c, icmd.Success)
	if !strings.Contains(res.String(), "nginx") {
		c.Fatalf("the search result should contain nginx")
	}

	resSpecificRegistry := command.PouchRun("search", "-r", "https://index.docker.io/v1/", "nginx")
	resSpecificRegistry.Assert(c, icmd.Success)
	if !strings.Contains(res.String(), "nginx") {
		c.Fatalf("the search result should contain nginx")
	}

	resWrongRegistry := command.PouchRun("search", "-r", "index.docker.io", "nginx")
	err := util.PartialEqual(resWrongRegistry.Stderr(), "unsupported protocol scheme")
	c.Assert(err, check.IsNil)
}
