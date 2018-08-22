package main

import (
	"encoding/json"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunWorkingDirSuite is the test suite for run CLI.
type PouchRunWorkingDirSuite struct{}

func init() {
	check.Suite(&PouchRunWorkingDirSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunWorkingDirSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunWorkingDirSuite) TearDownTest(c *check.C) {
}

// TestRunWithExistWorkingDir is to verify the valid running container
// with specifying working dir
func (suite *PouchRunWorkingDirSuite) TestRunWithExistWorkingDir(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)

	cname := "TestRunWithExistWorkingDir"
	res := command.PouchRun("run", "-d", "-w", "/root", "--name", cname, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].Config.WorkingDir, check.Equals, "/root")
}

// TestRunWithNotExistWorkingDir is to verify the valid running container
// with specifying a not exist working dir
func (suite *PouchRunWorkingDirSuite) TestRunWithNotExistWorkingDir(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)

	cname := "TestRunWithNotExistWorkingDir"
	res := command.PouchRun("run", "-d", "-w", "/tmp/notexist/dir", "--name", cname, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].Config.WorkingDir, check.Equals, "/tmp/notexist/dir")
}

// TestRunWithWorkingDir is to verify the valid running container
// with specifying working dir
func (suite *PouchRunWorkingDirSuite) TestRunWithWorkingDir(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)
	dir := "/tmp/testworkingdir"

	cname := "TestRunWithNotExistWorkingDir"
	res := command.PouchRun("run", "-w", dir, "--name", cname, busyboxImage, "pwd")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	out := res.Combined()
	out = strings.TrimSpace(out)
	if out != dir {
		c.Errorf("failed to set working directory, expect %s, got %s", dir, out)
	}
}

// TestRunWithWorkingDirExistAndIsFile is to verify the valid running container
// with specifying a exist file as working directory
func (suite *PouchRunWorkingDirSuite) TestRunWithWorkingDirExistAndIsFile(c *check.C) {
	SkipIfFalse(c, environment.IsMemorySupport)
	dir := "/bin/cat"

	cname := "TestRunWithWorkingDirExistAndIsFile"
	res := command.PouchRun("run", "-w", dir, "--name", cname, busyboxImage, "pwd")
	defer DelContainerForceMultyTime(c, cname)
	c.Assert(res.Stderr(), check.NotNil)

	expected := "not a directory"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Errorf("error information unmatched, expect %s, got %s", expected, out)
	}
}
