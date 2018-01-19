package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchPullSuite is the test suite fo help CLI.
type PouchPullSuite struct{}

func init() {
	check.Suite(&PouchPullSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchPullSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	c.Assert(environment.PruneAllImages(apiClient), check.IsNil)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchPullSuite) TearDownTest(c *check.C) {
	environment.PruneAllImages(apiClient)
}

// TestPullWorks tests "pouch pull" work.
func (suite *PouchPullSuite) TestPullWorks(c *check.C) {
	checkPull := func(target string, expected string) {
		command.PouchRun("pull", target).Assert(c, icmd.Success)
		res := command.PouchRun("images").Assert(c, icmd.Success)
		if out := res.Combined(); !strings.Contains(out, expected) {
			c.Fatalf("unexpected output %s: should got image %s\n", out, expected)
		}

		command.PouchRun("rmi", expected)
	}

	busybox := "registry.hub.docker.com/library/busybox"

	// without tag
	latest := busybox + ":latest"
	checkPull(busybox, latest)

	// with latest
	checkPull(latest, latest)

	// with :1.27.2
	version := busybox + ":1.27.2"
	checkPull(version, version)

	// without registry
	withoutRegistry := "busybox:latest"
	checkPull(withoutRegistry, latest)
}

// TestPullInWrongWay pulls in wrong way.
func (suite *PouchPullSuite) TestPullInWrongWay(c *check.C) {
	// pull unknown images
	{
		res := command.PouchRun("pull", "unknown")
		c.Assert(res.Error, check.NotNil)
	}

	// pull with invalid flag
	{
		res := command.PouchRun("pull", busyboxImage, "-f")
		c.Assert(res.Error, check.NotNil)
	}
}
