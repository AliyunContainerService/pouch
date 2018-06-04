package main

import (
	"fmt"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchWaitSuite is the test suite for wait CLI.
type PouchWaitSuite struct{}

func init() {
	check.Suite(&PouchWaitSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchWaitSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchWaitSuite) TearDownTest(c *check.C) {
}

// TestWaitNonBlockedExitZero is to verify the correctness of waiting a non-blocking container with 0 exit code
func (suite *PouchWaitSuite) TestWaitNonBlockedExitZero(c *check.C) {
	name := "TestWaitNonBlockedExitZero"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "sh", "-c", "true").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("wait", name)
	res.Assert(c, icmd.Success)
	output := res.Stdout()
	c.Assert(output, check.Equals, fmt.Sprintf("%s\n", "0"))
}

// TestWaitBlockedExitZero is to verify the correctness of waiting a blocking container with 0 exit code
func (suite *PouchWaitSuite) TestWaitBlockedExitZero(c *check.C) {
	name := "TestWaitBlockedExitZero"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "/bin/sh", "-c", "trap 'exit 0' TERM; "+
		"while true; do usleep 10; done").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	var output string
	chWait := make(chan struct{})
	go func() {
		chWait <- struct{}{}
		res := command.PouchRun("wait", name)
		res.Assert(c, icmd.Success)
		output = res.Stdout()
		close(chWait)
	}()
	<-chWait
	time.Sleep(100 * time.Millisecond)
	command.PouchRun("stop", name).Assert(c, icmd.Success)

	select {
	case <-chWait:
		c.Assert(output, check.Equals, fmt.Sprintf("%s\n", "0"))
	case <-time.After(2 * time.Second):
		c.Errorf("timeout waiting for `pouch wait` to exit")
	}
}

// TestWaitNonBlockedExitRandom is to verify the correctness of waiting a non-blocking container with random exit code
func (suite *PouchWaitSuite) TestWaitNonBlockedExitRandom(c *check.C) {
	name := "TestWaitNonBlockedExitRandom"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "sh", "-c", "exit 99").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("wait", name)
	res.Assert(c, icmd.Success)
	output := res.Stdout()
	c.Assert(output, check.Equals, fmt.Sprintf("%s\n", "99"))
}

// TestWaitBlockedExitRandom is to verify the correctness of waiting a blocking container with random exit code
func (suite *PouchWaitSuite) TestWaitBlockedExitRandom(c *check.C) {
	name := "TestWaitBlockedExitRandom"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "/bin/sh", "-c", "trap 'exit 99' TERM; "+
		"while true; do usleep 10; done").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	var output string
	chWait := make(chan struct{})
	go func() {
		chWait <- struct{}{}
		res := command.PouchRun("wait", name)
		res.Assert(c, icmd.Success)
		output = res.Stdout()
		close(chWait)
	}()
	<-chWait
	time.Sleep(100 * time.Millisecond)
	command.PouchRun("stop", name).Assert(c, icmd.Success)

	select {
	case <-chWait:
		c.Assert(output, check.Equals, fmt.Sprintf("%s\n", "99"))
	case <-time.After(2 * time.Second):
		c.Errorf("timeout waiting for `pouch wait` to exit")
	}
}
