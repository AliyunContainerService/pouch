package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchImagePruneSuite is the test suite for image prune CLI.
type PouchImagePruneSuite struct{}

func init() {
	check.Suite(&PouchImagePruneSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchImagePruneSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// PouchImagePruneSuite tests "pouch image prune" work.
func (suite *PouchImagePruneSuite) TestImagePruneWork(c *check.C) {
	PullImage(c, helloworldImage)
	PullImage(c, busyboxImage125)

	command.PouchRun("create", busyboxImage125, "top").Assert(c, icmd.Success)
	command.PouchRun("image", "prune", "-f").Assert(c, icmd.Success)

	res := command.PouchRun("images").Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, busyboxImage125) || strings.Contains(out, helloworldImage) {
		c.Fatalf("%s should contain image: %s and shouldn't contain image: %s\n", out, helloworldImage, busyboxImage125)
	}
}
