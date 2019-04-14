package main

import (
	"strings"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchStatsSuite is the test suite for stats CLI.
type PouchStatsSuite struct{}

func init() {
	check.Suite(&PouchStatsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (s *PouchStatsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (s *PouchStatsSuite) TearDownTest(c *check.C) {
}

func (s *PouchStatsSuite) TestStatsNoStream(c *check.C) {
	cname := "TestStatsNoStream"
	command.PouchRun("run", "-d", "--name", cname, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	cmd := command.PouchCmd("stats", "--no-stream", cname)
	res := icmd.StartCmd(cmd)
	res = icmd.WaitOnCmd(3*time.Second, res)
	res.Assert(c, icmd.Success)

	if !strings.Contains(res.Stdout(), cname) {
		c.Fatalf("container name not present in the stats output, %s", res.Stdout())
	}
}
