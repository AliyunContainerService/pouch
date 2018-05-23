package main

import (
	"encoding/json"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunUlimitSuite is the test suite for run CLI.
type PouchRunUlimitSuite struct{}

func init() {
	check.Suite(&PouchRunUlimitSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunUlimitSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunUlimitSuite) TearDownTest(c *check.C) {
}

// TestRunWithUlimit tests running container with --ulimit flag.
func (suite *PouchRunUlimitSuite) TestRunWithUlimit(c *check.C) {
	cname := "TestRunWithUlimit"
	res := command.PouchRun("run", "--ulimit", "nproc=256", "--name",
		cname, busyboxImage, "sh", "-c", "ulimit -p")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	out := res.Stdout()
	c.Assert(out, check.Equals, "256\n")

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	ul := result[0].HostConfig.Ulimits[0]
	c.Assert(ul.Name, check.Equals, "nproc")
	c.Assert(int(ul.Hard), check.Equals, 256)
	c.Assert(int(ul.Soft), check.Equals, 256)
}
