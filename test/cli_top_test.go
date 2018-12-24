package main

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchTopSuite is the test suite for top CLI.
type PouchTopSuite struct{}

func init() {
	check.Suite(&PouchTopSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchTopSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchTopSuite) TearDownTest(c *check.C) {
}

// TestTopStoppedContainer is to verify the correctness of top a stopped container.
func (suite *PouchTopSuite) TestTopStoppedContainer(c *check.C) {
	name := "TestTopStoppedContainer"

	res := command.PouchRun("create", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("top", name)
	c.Assert(res.Stderr(), check.NotNil)

	expectString := " is not running or paused, cannot execute top command"
	if out := res.Combined(); !strings.Contains(out, expectString) {
		// FIXME(ziren): for debug top error info is empty
		fmt.Printf("%+v", res)
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}
}

// TestTopRunningOrPausedContainer is to verify the correctness of pouch top command.
func (suite *PouchTopSuite) TestTopRunningOrPausedContainer(c *check.C) {
	name := "TestTopRunningOrPausedContainer"

	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("top", name)
	res.Assert(c, icmd.Success)

	expectString := "UIDPIDPPID"
	if out := util.TrimAllSpaceAndNewline(res.Combined()); !strings.HasPrefix(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}
	expectString = "top"
	if out := util.TrimAllSpaceAndNewline(res.Combined()); !strings.HasSuffix(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}

	// Top a paused contaner is valid
	res = command.PouchRun("pause", name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("top", name)
	res.Assert(c, icmd.Success)

	expectString = "UIDPIDPPID"
	if out := util.TrimAllSpaceAndNewline(res.Combined()); !strings.HasPrefix(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}
	expectString = "top"
	if out := util.TrimAllSpaceAndNewline(res.Combined()); !strings.HasSuffix(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}
}

// TestTopContainerWithOptions is to verify the correctness of pouch top command with ps options.
func (suite *PouchTopSuite) TestTopContainerWithOptions(c *check.C) {
	name := "TestTopContainerWithOptions"

	res := command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("top", name, "-aux")
	res.Assert(c, icmd.Success)

	expectString := "USERPID"
	if out := util.TrimAllSpaceAndNewline(res.Combined()); !strings.HasPrefix(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}
	expectString = "top"
	if out := util.TrimAllSpaceAndNewline(res.Combined()); !strings.HasSuffix(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}
}
