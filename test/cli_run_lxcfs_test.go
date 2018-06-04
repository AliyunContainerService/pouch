package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunLxcfsSuite is the test suite for run CLI.
type PouchRunLxcfsSuite struct{}

func init() {
	check.Suite(&PouchRunLxcfsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunLxcfsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunLxcfsSuite) TearDownTest(c *check.C) {
}

// TestRunEnableLxcfs is to verify run container with lxcfs.
func (suite *PouchRunLxcfsSuite) TestRunEnableLxcfs(c *check.C) {
	// TODO: also check if the pouchd started with lxcfs option
	SkipIfFalse(c, environment.IsLxcfsEnabled)
	name := "test-run-lxcfs"

	command.PouchRun("run", "-d", "--name", name,
		"-m", "512M", "--enableLxcfs=true",
		busyboxImage, "sleep", "10000").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("exec", name,
		"head", "-n", "5", "/proc/meminfo")
	res.Assert(c, icmd.Success)

	// the memory should be equal to 512M
	if out := res.Combined(); !strings.Contains(out, "524288 kB") {
		c.Fatalf("upexpected output %v, expected %s\n", res, "524288 kB")
	}
}
