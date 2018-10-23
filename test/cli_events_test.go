package main

import (
	"fmt"
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

func delEmptyStrInSlice(strSlice []string) []string {
	if len(strSlice) == 0 {
		return strSlice
	}

	newSlice := []string{}
	for _, v := range strSlice {
		if v != "" {
			newSlice = append(newSlice, v)
		}
	}

	return newSlice
}

// TestExecDieEventWorks tests exec_die event work.
func (suite *PouchEventsSuite) TestExecDieEventWorks(c *check.C) {
	name := "test-exec-die-event-works"

	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	// only works when test case run on the same machine with pouchd
	time.Sleep(1100 * time.Millisecond)
	start := time.Now()
	command.PouchRun("exec", name, "echo", "test").Assert(c, icmd.Success)
	time.Sleep(1100 * time.Millisecond)
	end := time.Now()

	since, until := start.Format(time.RFC3339), end.Format(time.RFC3339)
	res = command.PouchRun("events", "--since", since, "--until", until)
	output := res.Combined()

	// check output contains exec_die event
	lines := delEmptyStrInSlice(strings.Split(output, "\n"))
	if len(lines) != 1 {
		c.Errorf("unexpected output %s: should just contains 1 line", output)
	}

	if err := checkContainerEvent(lines[0], "exec_die"); err != nil {
		c.Errorf("exec_die event check error: %v", err)
	}
}

// TestDieEventWorks tests container die event work.
func (suite *PouchEventsSuite) TestDieEventWorks(c *check.C) {
	name := "test-die-event-works"

	res := command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	// only works when test case run on the same machine with pouchd
	time.Sleep(1100 * time.Millisecond)
	start := time.Now()
	command.PouchRun("stop", name).Assert(c, icmd.Success)
	time.Sleep(1100 * time.Millisecond)
	end := time.Now()

	since, until := start.Format(time.RFC3339), end.Format(time.RFC3339)
	res = command.PouchRun("events", "--since", since, "--until", until)
	output := res.Combined()

	// check events when stop a container
	lines := delEmptyStrInSlice(strings.Split(output, "\n"))
	if len(lines) != 2 {
		c.Errorf("unexpected output %s: should contains 2 event line when stop a container", output)
	}

	if err := checkContainerEvent(lines[0], "die"); err != nil {
		c.Errorf("die event check error: %v", err)
	}

	if err := checkContainerEvent(lines[1], "stop"); err != nil {
		c.Errorf("exec_die event check error: %v", err)
	}
}

func checkContainerEvent(eventStr, eventType string) error {
	strSlice := strings.Split(eventStr, " ")
	if len(strSlice) < 4 {
		return fmt.Errorf("unexpected output %s: output line may not be container event", eventStr)
	}

	if strSlice[1] != "container" || strSlice[2] != eventType {
		return fmt.Errorf("unexpected output %s: should be %s events", eventStr, eventType)
	}

	return nil
}
