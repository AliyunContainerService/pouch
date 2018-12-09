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

// TestRmiForce tests "pouch rmi -f" work
func (suite *PouchRmiSuite) TestRmiForce(c *check.C) {
	command.PouchRun("pull", helloworldImage).Assert(c, icmd.Success)

	// TODO: rmi -f after create/start containers.
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
