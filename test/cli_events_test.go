package main

import (
	"strings"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/gotestyourself/gotestyourself/icmd"

	"github.com/go-check/check"
)

// PouchEventsSuite is the test suite for events CLI.
type PouchEventsSuite struct{}

func init() {
	check.Suite(&PouchEventsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchEventsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchEventsSuite) TearDownTest(c *check.C) {

}

// TestEventsWorks tests "pouch events" work.
func (suite *PouchEventsSuite) TestEventsWorks(c *check.C) {
	name := "test-events-works"

	// only works when test case run on the same machine with pouchd
	start := time.Now()
	time.Sleep(1100 * time.Millisecond)
	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)
	time.Sleep(1100 * time.Millisecond)
	end := time.Now()

	since, until := start.Format(time.RFC3339), end.Format(time.RFC3339)
	res = command.PouchRun("events", "--since", since, "--until", until)
	if out := res.Combined(); !(strings.Contains(out, "create") && strings.Contains(out, "start")) {
		c.Errorf("unexpected output %s: should contains create and start events\n", out)
	}
}
