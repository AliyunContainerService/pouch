package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchPsSuite is the test suite fo help CLI.
type PouchPsSuite struct{}

func init() {
	check.Suite(&PouchPsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchPsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	c.Assert(environment.PruneAllContainers(apiClient), check.IsNil)
	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchPsSuite) TearDownTest(c *check.C) {
	environment.PruneAllContainers(apiClient)
}

// TestPsWorks tests "pouch ps" work.
//
// TODO: check more value, like id/runtime.
func (suite *PouchPsSuite) TestPsWorks(c *check.C) {
	name := "ps-normal"

	// create
	{
		command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)

		res := command.PouchRun("ps", "-a").Assert(c, icmd.Success)
		kv := psToKV(res.Combined())

		c.Assert(kv[name].status[0], check.Equals, "created")
		c.Assert(kv[name].image, check.Equals, busyboxImage)
	}

	// running
	{
		command.PouchRun("start", name).Assert(c, icmd.Success)

		res := command.PouchRun("ps").Assert(c, icmd.Success)
		kv := psToKV(res.Combined())

		c.Assert(kv[name].status[0], check.Equals, "Up")
	}

	// stop
	{
		command.PouchRun("stop", name).Assert(c, icmd.Success)

		res := command.PouchRun("ps", "-a").Assert(c, icmd.Success)
		kv := psToKV(res.Combined())

		c.Assert(kv[name].status[0], check.Equals, "stopped")
	}
}

// psTable represents the table of "pouch ps" result.
type psTable struct {
	id      string
	name    string
	status  []string
	created []string
	image   string
	runtime string
}

// psToKV parse "pouch ps" into key-value mapping.
func psToKV(ps string) map[string]psTable {
	// skip header
	lines := strings.Split(ps, "\n")[1:]

	res := make(map[string]psTable)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		items := strings.Fields(line)

		pst := psTable{}
		pst.name = items[0]
		pst.id = items[1]

		if items[2] == "Up" {
			pst.status = items[2:5]
			pst.created = items[5:8]
			pst.image = items[8]
			pst.runtime = items[9]
		} else {
			pst.status = items[2:3]
			pst.created = items[3:6]
			pst.image = items[6]
			pst.runtime = items[7]
		}
		res[items[0]] = pst
	}
	return res
}
