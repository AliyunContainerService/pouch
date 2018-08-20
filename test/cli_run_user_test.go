package main

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunUserSuite is the test suite for run CLI.
type PouchRunUserSuite struct{}

func init() {
	check.Suite(&PouchRunUserSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunUserSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunUserSuite) TearDownTest(c *check.C) {
}

// TestRunWithUser is to verify run container with user.
func (suite *PouchRunUserSuite) TestRunWithUser(c *check.C) {
	namePrefix := "run-user"

	for idx, user := range []string{
		"1001", "1001:1001", "root", "root:root", "1001:root", "root:1001",
	} {
		name := fmt.Sprintf("%s-%d", namePrefix, idx)
		res := command.PouchRun("run", "-d", "--name", name,
			"--user", user, busyboxImage, "top")
		res.Assert(c, icmd.Success)
		DelContainerForceMultyTime(c, name)
	}
}

// TestRunWithUserFail is to verify run container with wrong user will fails.
func (suite *PouchRunUserSuite) TestRunWithUserFail(c *check.C) {
	namePrefix := "run-user-fail"

	for idx, user := range []string{
		"wrong-user", "wrong-user:wrong-group", "1001:wrong-group", "wrong-user:1001",
	} {
		name := fmt.Sprintf("%s-%d", namePrefix, idx)
		res := command.PouchRun("run", "-d", "--name", name,
			"--user", user, busyboxImage, "top")
		c.Assert(res.Stderr(), check.NotNil)
		DelContainerForceMultyTime(c, name)
	}
}

// TestRunWithUser is to verify run container with user.
func (suite *PouchRunUserSuite) TestRunWithAddUser(c *check.C) {
	name := "run-user-admin"
	{
		res := command.PouchRun("run", "-d", "--name", name,
			busyboxImage, "top")
		defer DelContainerForceMultyTime(c, name)
		res.Assert(c, icmd.Success)
	}

	{
		res := command.PouchRun("exec", name, "adduser",
			"--disabled-password", "admin")
		res.Assert(c, icmd.Success)
	}

	{
		res := command.PouchRun("exec", "-u", "admin", name, "whoami")
		res.Assert(c, icmd.Success)
		if !strings.Contains(res.Stdout(), "admin") {
			c.Errorf("failed to start a container with user: %s",
				res.Stdout())
		}
	}
}
