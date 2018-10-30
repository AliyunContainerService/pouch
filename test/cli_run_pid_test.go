package main

import (
	"os"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunPidSuite is the test suite for run CLI.
type PouchRunPidSuite struct{}

func init() {
	check.Suite(&PouchRunPidSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunPidSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunPidSuite) TearDownTest(c *check.C) {
}

// TestRunWithPIDMode is to verify --specific PID mode when running a container.
// TODO: test container pid namespace mode.
func (suite *PouchRunPidSuite) TestRunWithPIDMode(c *check.C) {
	name := "test-run-with-pid-mode"

	res := command.PouchRun("run", "-d", "--name", name,
		"--pid", "host", busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)
}

// TestRunWithPidsLimit tests running container with --pids-limit flag.
func (suite *PouchRunPidSuite) TestRunWithPidsLimit(c *check.C) {
	// pids cgroup may not supported in inner ci
	SkipIfFalse(c, func() bool {
		if _, err := os.Stat("/sys/fs/cgroup/pids"); err != nil {
			return false
		}
		return true
	})

	cname := "TestRunWithPidsLimit"
	pidfile := "/sys/fs/cgroup/pids/pids.max"
	res := command.PouchRun("run", "--pids-limit", "10",
		"--name", cname, busyboxImage, "cat", pidfile)
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	out := res.Stdout()
	c.Assert(out, check.Equals, "10\n")

	pidsLimit, err := inspectFilter(cname, ".HostConfig.PidsLimit")
	c.Assert(err, check.IsNil)
	c.Assert(pidsLimit, check.Equals, "10")
}
