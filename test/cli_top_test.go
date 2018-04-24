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

	command.PouchRun("create", "-m", "300M", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("top", name)
	c.Assert(res.Error, check.NotNil)

	expectString := "container is not running, can not execute top command"
	if out := res.Combined(); !strings.Contains(out, expectString) {
		// FIXME(ziren): for debug top error info is empty
		fmt.Printf("%+v", res)
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestTopContainer is to verify the correctness of pouch top command.
func (suite *PouchTopSuite) TestTopContainer(c *check.C) {
	name := "TestTopContainer"

	command.PouchRun("run", "-d", "-m", "300M", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)

	res := command.PouchRun("top", name)
	c.Assert(res.Error, check.IsNil)

	expectString := "UIDPIDPPID"
	if out := util.TrimAllSpaceAndNewline(res.Combined()); !strings.Contains(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}
