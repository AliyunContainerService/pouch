package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchInspectSuite is the test suite for inspect CLI.
type PouchInspectSuite struct{}

func init() {
	check.Suite(&PouchInspectSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchInspectSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchInspectSuite) TearDownTest(c *check.C) {
}

// TestInspectFormat is to verify the format flag of inspect command.
func (suite *PouchInspectSuite) TestInspectFormat(c *check.C) {
	name := "inspect-format-print"

	command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result.ID

	// inspect Container ID
	output = command.PouchRun("inspect", "-f", "{{.ID}}", name).Stdout()
	c.Assert(output, check.Equals, fmt.Sprintf("%s\n", containerID))

	// inspect Memory
	output = command.PouchRun("inspect", "-f", "{{.HostConfig.Memory}}", name).Stdout()
	c.Assert(output, check.Equals, fmt.Sprintf("%d\n", result.HostConfig.Memory))

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)
}

// TestInspectWrongFormat is to verify using wrong format flag of inspect command.
func (suite *PouchInspectSuite) TestInspectWrongFormat(c *check.C) {
	name := "inspect-wrong-format-print"

	command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("inspect", "-f", "{{.NotExists}}", name)
	c.Assert(res.Error, check.NotNil)

	expectString := "Template parsing error"
	if out := res.Combined(); !strings.Contains(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}

	command.PouchRun("rm", "-f", name).Assert(c, icmd.Success)

}
