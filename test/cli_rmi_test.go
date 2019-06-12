package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRmiSuite is the test suite for rmi CLI.
type PouchRmiSuite struct{}

func init() {
	check.Suite(&PouchRmiSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRmiSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	environment.PruneAllContainers(apiClient)
}

// TestRmiWorks tests "pouch rmi" work.
func (suite *PouchRmiSuite) TestRmiWorks(c *check.C) {
	command.PouchRun("pull", helloworldImage).Assert(c, icmd.Success)

	command.PouchRun("rmi", helloworldImage).Assert(c, icmd.Success)

	res := command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); strings.Contains(out, helloworldImage) {
		c.Fatalf("unexpected output %s: should rm image %s\n", out, helloworldImage)
	}
}

// TestRmiWithRunningContainer tests "pouch rmi" won't work without force
func (suite *PouchRmiSuite) TestRmiWithRunningContainer(c *check.C) {
	name := "TestRmiWithRunningContainer"
	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// Can't remove without "--force"
	command.PouchRun("rmi", busyboxImage)
	res := command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, busyboxImage) {
		c.Fatalf("unexpected output %s: shouldn't rm image %s without force\n", out, busyboxImage)
	}

	command.PouchRun("rmi", "-f", busyboxImage).Assert(c, icmd.Success)
	res = command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); strings.Contains(out, busyboxImage) {
		c.Fatalf("unexpected output %s: should rm image %s\n", out, busyboxImage)
	}
}

// TestRmiForce tests "pouch rmi -f" work
func (suite *PouchRmiSuite) TestRmiForce(c *check.C) {
	command.PouchRun("pull", helloworldImage).Assert(c, icmd.Success)

	command.PouchRun("rmi", "-f", helloworldImage).Assert(c, icmd.Success)

	res := command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); strings.Contains(out, helloworldImage) {
		c.Fatalf("unexpected output %s: should rm image %s\n", out, helloworldImage)
	}
}

// TestRmiByImageID tests "pouch rmi {ID}" work.
func (suite *PouchRmiSuite) TestRmiByImageID(c *check.C) {
	command.PouchRun("pull", helloworldImage).Assert(c, icmd.Success)

	res := command.PouchRun("images")
	res.Assert(c, icmd.Success)
	imageID := imagesListToKV(res.Combined())[helloworldImage][0]

	command.PouchRun("rmi", imageID).Assert(c, icmd.Success)

	res = command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); strings.Contains(out, helloworldImage) {
		c.Fatalf("unexpected output %s: should rm image %s\n", out, helloworldImage)
	}
}

// TestRmiByImageDigestID tests "pouch rmi sha256:xxx" work.
func (suite *PouchRmiSuite) TestRmiByImageDigestID(c *check.C) {
	command.PouchRun("pull", helloworldImage).Assert(c, icmd.Success)

	res := command.PouchRun("images")
	res.Assert(c, icmd.Success)
	imageID := imagesListToKV(res.Combined())[helloworldImage][0]

	command.PouchRun("rmi", "sha256:"+imageID).Assert(c, icmd.Success)

	res = command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); strings.Contains(out, helloworldImage) {
		c.Fatalf("unexpected output %s: should rm image %s\n", out, helloworldImage)
	}
}

// TestRmiByImageIDWithTwoPrimaryReferences tests "pouch rmi {ID}" work.
func (suite *PouchRmiSuite) TestRmiByImageIDWithTwoPrimaryReferences(c *check.C) {
	var (
		repoTag    = environment.BusyboxRepo + ":" + environment.BusyboxTag
		repoDigest = environment.BusyboxRepo + "@" + environment.BusyboxDigest
	)

	command.PouchRun("pull", repoTag).Assert(c, icmd.Success)
	command.PouchRun("pull", repoDigest).Assert(c, icmd.Success)

	res := command.PouchRun("images")
	res.Assert(c, icmd.Success)
	imageID := imagesListToKV(res.Combined())[repoTag][0]

	command.PouchRun("rmi", imageID).Assert(c, icmd.Success)

	res = command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); strings.Contains(out, repoTag) {
		c.Fatalf("unexpected output %s: should rm image %s\n", out, repoTag)
	}
}

// TestRmiByImageIDWithPrimaryReferencesAndRunningContainers tests "pouch rmi {ID}" won't work without force.
func (suite *PouchRmiSuite) TestRmiByImageIDWithPrimaryReferencesAndRunningContainers(c *check.C) {
	busyboxImageAlias := busyboxImage + "Alias"
	name := "TestRmiByImageIDWithRunningContainer"

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
	command.PouchRun("tag", busyboxImage, busyboxImageAlias).Assert(c, icmd.Success)
	res := command.PouchRun("images")
	res.Assert(c, icmd.Success)
	imageID := imagesListToKV(res.Combined())[busyboxImage][0]

	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// Can't remove without "--force"
	res = command.PouchRun("rmi", imageID)
	c.Assert(res.Stderr(), check.NotNil, check.Commentf("Unable to remove the image"))
	res = command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, busyboxImage) {
		c.Fatalf("unexpected output %s: shouldn't rm image %s without force\n", out, busyboxImage)
	}

	command.PouchRun("rmi", "-f", imageID).Assert(c, icmd.Success)
	res = command.PouchRun("images").Assert(c, icmd.Success)
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
		c.Assert(res.Stderr(), check.NotNil, check.Commentf(tc.name))
	}
}
