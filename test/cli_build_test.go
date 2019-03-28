package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchBuildSuite is the test suite for container create API.
type PouchBuildSuite struct{}

func init() {
	check.Suite(&PouchBuildSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchBuildSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestBuildMultipleStage tests create checkpoint
func (suite *PouchBuildSuite) TestBuildMultipleStage(c *check.C) {
	iname := fmt.Sprintf("%s:%v", c.TestName(), time.Now().UnixNano())

	path := filepath.Join("testdata", "build", "multiplestage")
	command.PouchRun("build", path, "-t", iname).Assert(c, icmd.Success)
	defer command.PouchRun("rmi", iname)

	res := command.PouchRun("run", "--name", c.TestName(), iname)
	defer DelContainerForceMultyTime(c, c.TestName())
	res.Assert(c, icmd.Success)
	c.Assert(res.Stdout(), check.Equals, "Hi PouchContainer!\n")
}
