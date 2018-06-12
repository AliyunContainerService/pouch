package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchLogsSuite is the test suite for logs CLI.
type PouchLogsSuite struct{}

func init() {
	check.Suite(&PouchLogsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchLogsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestCreatedContainerLogIsEmpty tests logs for created container.
func (suite *PouchLogsSuite) TestCreatedContainerLogIsEmpty(c *check.C) {
	cname := "TestCLILogs_EmptyLogInCreatedContainer"

	command.PouchRun("create", "--name", cname, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	res := command.PouchRun("logs", cname)
	res.Assert(c, icmd.Success)
	c.Assert(res.Combined(), check.Equals, "")
}

// TestSinceAndUntil tests the since and until.
func (suite *PouchLogsSuite) TestSinceAndUntil(c *check.C) {
	cname := "TestCLILogs_Since_and_Until"
	totalLine := 5

	command.PouchRun(
		"run",
		"-t",
		"--name", cname, busyboxImage,
		"sh", "-c", fmt.Sprintf("for i in $(seq 1 %v); do echo hello$i; sleep 1; done;", totalLine),
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	allLogs := suite.syncLogs(c, cname, "--timestamps")
	c.Assert(len(allLogs), check.Equals, totalLine)

	// get the since and until time
	sinceTime, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[2], " ")[0])
	c.Assert(err, check.IsNil)

	untilTime, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[3], " ")[0])
	c.Assert(err, check.IsNil)
	untilTime = untilTime.Add(-200 * time.Nanosecond)

	allLogs = suite.syncLogs(c, cname,
		"--since", fmt.Sprintf("%d", sinceTime.UnixNano()),
		"--until", fmt.Sprintf("%d", untilTime.UnixNano()))
	c.Assert(len(allLogs), check.Equals, 1)
}

// TestTimestamp tests the timestamps flag.
func (suite *PouchLogsSuite) TestTimestamp(c *check.C) {
	cname := "TestCLILogs_timestamp"

	command.PouchRun(
		"run",
		"-t",
		"--name", cname,
		busyboxImage,
		"echo", "hello",
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	allLogs := suite.syncLogs(c, cname, "--timestamps")
	c.Assert(len(allLogs), check.Equals, 1)

	_, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[0], " ")[0])
	c.Assert(err, check.IsNil)
	c.Assert(strings.TrimSpace(strings.Split(allLogs[0], " ")[1]), check.Equals, "hello")
}

// TestTailMode tests follow mode.
func (suite *PouchLogsSuite) TestTailLine(c *check.C) {
	cname := "TestCLILogs_tail_line"
	DelContainerForceMultyTime(c, cname)
	totalLine := 100

	command.PouchRun(
		"run",
		"-t",
		"--name", cname,
		busyboxImage,
		"sh", "-c", fmt.Sprintf("for i in $(seq 1 %v); do echo hello-$i; done;", totalLine),
	).Assert(c, icmd.Success)

	for _, tc := range []struct {
		input    string
		expected int
	}{
		{"1000", totalLine},
		{"100", totalLine},
		{"67", 67},
		{"5", 5},
		{"0", totalLine},
		{"-1", totalLine},
		{"wronglinenumber", totalLine},
	} {
		allLogs := suite.syncLogs(c, cname, "--tail", tc.input)
		c.Assert(allLogs, check.HasLen, tc.expected)
	}
}

// TestFollowMode tests follow mode.
func (suite *PouchLogsSuite) TestFollowMode(c *check.C) {
	cname := "TestCLILogs_follow_mode"

	command.PouchRun(
		"run",
		"-d",
		"--name", cname,
		busyboxImage,
		"sh", "-c", "for i in $(seq 1 3); do sleep 2; echo hello; done;",
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	waitCh := make(chan error)
	go func() {
		waitCh <- command.PouchRun("logs", "-f", cname).Error
		close(waitCh)
	}()

	select {
	case err := <-waitCh:
		c.Assert(err, check.IsNil)
	case <-time.After(15 * time.Second):
		c.Fatal("pouch logs -f may hanged")
	}
}

func (suite *PouchLogsSuite) syncLogs(c *check.C, cname string, flags ...string) []string {
	args := append([]string{"logs"}, flags...)

	res := command.PouchRun(append(args, cname)...)
	res.Assert(c, icmd.Success)

	return strings.Split(strings.TrimSpace(string(res.Combined())), "\n")
}
