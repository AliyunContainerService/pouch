package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunPrivilegedSuite is the test suite for run CLI.
type PouchRunPrivilegedSuite struct{}

func init() {
	check.Suite(&PouchRunPrivilegedSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunPrivilegedSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunPrivilegedSuite) TearDownTest(c *check.C) {
}

// TestRunWithAndWithoutPrivileged is to verify run container with privilege.
func (suite *PouchRunPrivilegedSuite) TestRunWithAndWithoutPrivileged(c *check.C) {
	name := "TestRunWithPrivileged"
	command.PouchRun("run", "--name", name, "--privileged", busyboxImage, "brctl", "addbr", "foobar").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	name1 := "TestRunWithoutPrivileged"
	res := command.PouchRun("run", "--name", name1, busyboxImage, "brctl", "addbr", "foobar")
	defer DelContainerForceMultyTime(c, name1)
	if res.ExitCode == 0 {
		c.Errorf("non-privileged container executes brctl should failed, but succeeded: %v", res.Combined())
	}

	expected := "Operation not permitted"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Errorf("expected %s, but got %s", expected, out)
	}
}

func (suite *PouchRunPrivilegedSuite) TestRunCheckProcWritableWithAndWithoutPrivileged(c *check.C) {
	name := "TestRunProcWritableInPrivileged"
	command.PouchRun("run", "--name", name, "--privileged", busyboxImage, "sh", "-c", "touch /proc/sysrq-trigger").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	name1 := "TestRunProcNotWritableInNonPrivileged"
	res := command.PouchRun("run", "--name", name1, busyboxImage, "sh", "-c", "touch /proc/sysrq-trigger")
	defer DelContainerForceMultyTime(c, name1)

	if res.ExitCode == 0 {
		c.Errorf("non-privileged container executes touch /proc/sysrq-trigger should failed, but succeeded: %v", res.Combined())
	}

	expected := "Read-only file system"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Errorf("expected %s, but got %s", expected, out)
	}
}

func (suite *PouchRunPrivilegedSuite) TestRunCheckSysWritableWithAndWithoutPrivileged(c *check.C) {
	name := "TestRunSysWritableInPrivileged"
	command.PouchRun("run", "--name", name, "--privileged", busyboxImage, "sh", "-c", "touch /sys/kernel/profiling").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	name1 := "TestRunSysNotWritableInNonPrivileged"
	res := command.PouchRun("run", "--name", name1, busyboxImage, "sh", "-c", "touch /sys/kernel/profiling")
	defer DelContainerForceMultyTime(c, name1)

	if res.ExitCode == 0 {
		c.Errorf("non-privileged container executes touch /sys/kernel/profiling should failed, but succeeded: %v", res.Combined())
	}

	expected := "Read-only file system"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Errorf("expected %s, but got %s", expected, out)
	}
}
