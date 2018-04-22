package main

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchPsSuite is the test suite for ps CLI.
type PouchPsSuite struct{}

func init() {
	check.Suite(&PouchPsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchPsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)
	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchPsSuite) TearDownTest(c *check.C) {
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

		c.Assert(kv[name].status[0], check.Equals, "Stopped")
	}

	defer DelContainerForceMultyTime(c, name)
}

// TestPsAll tests "pouch ps -a" work
func (suite *PouchPsSuite) TestPsAll(c *check.C) {
	name := "ps-all"

	command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("ps").Assert(c, icmd.Success)
	lines := strings.Split(res.Combined(), "\n")

	// show running containers default
	c.Assert(lines[1], check.Equals, "")

	res = command.PouchRun("ps", "-a").Assert(c, icmd.Success)
	kv := psToKV(res.Combined())

	c.Assert(kv[name].status[0], check.Equals, "created")
}

// TestPsQuiet tests "pouch ps -q" work
func (suite *PouchPsSuite) TestPsQuiet(c *check.C) {
	name := "ps-quiet"

	command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	res := command.PouchRun("ps", "-q", "-a").Assert(c, icmd.Success)
	lines := strings.Split(res.Combined(), "\n")

	for _, line := range lines {
		if line != "" {
			match, _ := regexp.MatchString("^[0-9a-f]{6}$", line)
			c.Assert(match, check.Equals, true)
		}
	}
}

// TestPsNoTrunc tests "pouch ps trunc" work
func (suite *PouchPsSuite) TestPsNoTrunc(c *check.C) {
	name := "ps-noTrunc"

	command.PouchRun("create", "--name", name, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	command.PouchRun("start", name).Assert(c, icmd.Success)

	res := command.PouchRun("ps", "--no-trunc").Assert(c, icmd.Success)
	kv := psToKV(res.Combined())

	// Use inspect command to get container id
	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(kv[name].id, check.HasLen, 64)
	c.Assert(kv[name].id, check.Equals, result[0].ID)
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
		} else if items[2] == "Stopped" || items[2] == "Exited" {
			pst.status = items[2:6]
			pst.created = items[6:9]
			pst.image = items[9]
			pst.runtime = items[10]
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
