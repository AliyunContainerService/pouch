package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
)

// PouchRunDNSSuite is the test suite for run CLI.
type PouchRunEnvfileSuite struct{}

func init() {
	check.Suite(&PouchRunEnvfileSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunEnvfileSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunEnvfileSuite) TearDownTest(c *check.C) {
}

func (suite *PouchRunDNSSuite) TestRunWithEnvfileVariables(c *check.C) {
	/*
		cname := "TestRunWithEnvfileVariables"

		e := "open nonexistent: no such file or directory"
		res := command.PouchRun("run","--name ",cname, "--env-file=nonexistent", "-d","busybox", "top")
		res.Assert(c, icmd.Success)
		c.Assert(res.Combined(),check.Equals,e)
		res = command.PouchRun("run","--name ",cname, "--env-file=fixtures/valid.env","-d", "busybox", "top")
		res.Assert(c, icmd.Success)
		res=command.PouchRun("exec",cname,"env | grep ENV1")
		res.Assert(c, icmd.Success)
		c.Assert(res.Combined(),check.Equals,"ENV1=value1")
		res = command.PouchRun("run","--name ",cname, "--env-file=fixtures/valid.env","--env=ENV2=value2", "-d","busybox", "top")
		res.Assert(c, icmd.Success)
		res=command.PouchRun("exec",cname,"env | grep ENV")
		res.Assert(c, icmd.Success)
		c.Assert(res.Combined(),check.Equals,"ENV1=value1\nENV2=value2")

	*/
}
