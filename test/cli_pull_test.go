package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchPullSuite is the test suite for pull CLI.
type PouchPullSuite struct{}

func init() {
	check.Suite(&PouchPullSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchPullSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchPullSuite) TearDownTest(c *check.C) {
}

// TestPullWorks tests "pouch pull" work.
func (suite *PouchPullSuite) TestPullWorks(c *check.C) {
	checkPull := func(target string, expected string) {
		command.PouchRun("pull", target).Assert(c, icmd.Success)
		res := command.PouchRun("images").Assert(c, icmd.Success)
		if out := res.Combined(); !strings.Contains(out, expected) {
			c.Fatalf("unexpected output %s: should got image %s\n", out, expected)
		}

		command.PouchRun("rmi", "-f", expected).Assert(c, icmd.Success)
	}

	busyboxRepo := environment.BusyboxRepo

	// without tag
	latest := busyboxRepo + ":latest"
	checkPull(busyboxRepo, latest)

	// with latest
	checkPull(latest, latest)

	// with default tag
	version := busyboxRepo + ":" + environment.BusyboxTag
	checkPull(version, version)

	// image with digest for tag
	busyboxDigest := busyboxRepo + "@" + environment.BusyboxDigest
	busyboxDigestWithWrongTag := busyboxRepo + ":whatever" + "@" + environment.BusyboxDigest
	checkPull(busyboxDigestWithWrongTag, busyboxDigest)
}

// TestPullInWrongWay pulls in wrong way.
func (suite *PouchPullSuite) TestPullInWrongWay(c *check.C) {
	// pull unknown images
	{
		res := command.PouchRun("pull", "unknown")
		c.Assert(res.Stderr(), check.NotNil)
	}

	// pull with invalid flag
	{
		res := command.PouchRun("pull", busyboxImage, "-f")
		c.Assert(res.Stderr(), check.NotNil)
	}
}
