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

		command.PouchRun("rmi", expected).Assert(c, icmd.Success)
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

	// image with namespace but without registry
	cadvisor := "registry.hub.docker.com/google/cadvisor:latest"
	cadvisorWithoutRegistry := "google/cadvisor:latest"
	checkPull(cadvisorWithoutRegistry, cadvisor)

	// image with digest for tag 1.25
	busybox125Digest := "sha256:29f5d56d12684887bdfa50dcd29fc31eea4aaf4ad3bec43daf19026a7ce69912"

	busyboxDigest := busybox + "@" + busybox125Digest
	busyboxDigestWithWrongTag := busybox + ":whatever" + "@" + busybox125Digest
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
