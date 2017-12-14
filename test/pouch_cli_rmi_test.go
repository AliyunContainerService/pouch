package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRmiSuite is the test suite fo help CLI.
type PouchRmiSuite struct{}

func init() {
	check.Suite(&PouchRmiSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRmiSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestRmiWorks tests "pouch rmi" work.
func (suite *PouchRmiSuite) TestRmiWorks(c *check.C) {
	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)

	command.PouchRun("rmi", busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); strings.Contains(out, busyboxImage) {
		c.Fatalf("unexpected output %s: should rm image %s\n", out, busyboxImage)
	}
}

// TestRmiForce tests "pouch rmi -f" work
func (suite *PouchRmiSuite) TestRmiForce(c *check.C) {
	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)

	// TODO: rmi -f after create/start containers.
	command.PouchRun("rmi", "-f", busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); strings.Contains(out, busyboxImage) {
		c.Fatalf("unexpected output %s: should rm image %s\n", out, busyboxImage)
	}
}

// TestRmiInWrongWay run rmi in wrong ways.
func (suite *PouchRmiSuite) TestRmiInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "unknown flag", args: "-a"},
		{name: "unknown image name", args: "unknown"},

		// TODO: should add the following cases if ready
		// {name: "missing image name", args: ""},
	} {
		res := command.PouchRun("rmi", tc.args)
		c.Assert(res.Error, check.NotNil, check.Commentf(tc.name))
	}
}
