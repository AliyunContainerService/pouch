package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

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

	c.Assert(util.PartialEqual(res.Combined(), "Operation not permitted"), check.IsNil)
}

func (suite *PouchRunPrivilegedSuite) TestRunCheckProcWritableWithAndWithoutPrivileged(c *check.C) {
	name := "TestRunProcWritableInPrivileged"
	command.PouchRun("run", "--name", name, "--privileged", busyboxImage, "sh", "-c", "touch /proc/sysrq-trigger").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	name1 := "TestRunProcNotWritableInNonPrivileged"
	res := command.PouchRun("run", "--name", name1, busyboxImage, "sh", "-c", "touch /proc/sysrq-trigger")
	defer DelContainerForceMultyTime(c, name1)

	c.Assert(util.PartialEqual(res.Combined(), "Read-only file system"), check.IsNil)
}

func (suite *PouchRunPrivilegedSuite) TestRunCheckSysWritableWithAndWithoutPrivileged(c *check.C) {
	name := "TestRunSysWritableInPrivileged"
	command.PouchRun("run", "--name", name, "--privileged", busyboxImage, "sh", "-c", "touch /sys/kernel/profiling").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	name1 := "TestRunSysNotWritableInNonPrivileged"
	res := command.PouchRun("run", "--name", name1, busyboxImage, "sh", "-c", "touch /sys/kernel/profiling")
	defer DelContainerForceMultyTime(c, name1)

	c.Assert(util.PartialEqual(res.Combined(), "Read-only file system"), check.IsNil)
}

// TestCgroupWritableWithAndWithoutPrivileged tests cgroup can be writable with privileged,
// can not be writable without privileged
func (suite *PouchRunPrivilegedSuite) TestCgroupWritableWithAndWithoutPrivileged(c *check.C) {
	name := "TestRunCheckCgroupWritable"
	command.PouchRun("run", "--name", name, "--privileged", busyboxImage, "sh", "-c", "mkdir /sys/fs/cgroup/cpu/test").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	name1 := "TestRunCheckCgroupCannotWritable"
	res := command.PouchRun("run", "--name", name1, busyboxImage, "sh", "-c", "mkdir /sys/fs/cgroup/cpu/test")
	defer DelContainerForceMultyTime(c, name1)

	c.Assert(util.PartialEqual(res.Combined(), "Read-only file system"), check.IsNil)
}
