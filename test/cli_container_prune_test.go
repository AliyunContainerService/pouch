package main

import (
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchContainerPruneSuite is the test suite for container prune.
type PouchContainerPruneSuite struct{}

func init() {
	check.Suite(&PouchContainerPruneSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchContainerPruneSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

}

// TestContainerPruneWork tests "pouch container prune" work.
func (suite *PouchContainerPruneSuite) TestContainerPruneWork(c *check.C) {
	name := "container_prune_work"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
	command.PouchRun("stop", name).Assert(c, icmd.Success)
	command.PouchRun("container", "prune", "-f").Assert(c, icmd.Success)

	res := command.PouchRun("ps", "-a").Assert(c, icmd.Success)
	kv := psToKV(res.Combined())
	_, exist := kv[name]
	c.Assert(exist, check.Equals, false)
}

// TestPsFilterInvalid tests "pouch ps -f" invalid
func (suite *PouchContainerPruneSuite) TestContainerPruneFilter(c *check.C) {
	result := command.PouchRun("container", "prune", "-f", "--filter", "aaa")
	err := util.PartialEqual(result.Stderr(), "Bad format of filter, expected name=value")
	c.Assert(err, check.IsNil)

	result = command.PouchRun("container", "prune", "-f", "--filter", "a=b")
	err = util.PartialEqual(result.Stderr(), "Invalid filter")
	c.Assert(err, check.IsNil)

	nameA := "equal-name-a"
	command.PouchRun("run", "-d", "--name", nameA, "-l", "run=a", busyboxImage, "top").Assert(c, icmd.Success)
	command.PouchRun("stop", nameA).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, nameA)

	nameB := "equal-name-b"
	command.PouchRun("run", "-d", "--name", nameB, "-l", "run=b", busyboxImage, "top").Assert(c, icmd.Success)
	command.PouchRun("stop", nameB).Assert(c, icmd.Success)

	command.PouchRun("container", "prune", "-f", "--filter", "label=run=b").Assert(c, icmd.Success)

	res := command.PouchRun("ps", "-a").Assert(c, icmd.Success)
	kv := psToKV(res.Combined())
	_, exist1 := kv[nameA]
	_, exist2 := kv[nameB]
	c.Assert(exist1, check.Equals, true)
	c.Assert(exist2, check.Equals, false)
}
