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

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchInspectSuite) TearDownTest(c *check.C) {
}

// TestInspectFormat is to verify the format flag of inspect command.
func (suite *PouchInspectSuite) TestInspectFormat(c *check.C) {
	name := "inspect-format-print"

	res := command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	containerID := result[0].ID

	// inspect Container ID
	output = command.PouchRun("inspect", "-f", "{{.ID}}", name).Stdout()
	c.Assert(output, check.Equals, fmt.Sprintf("%s\n", containerID))

	// inspect Memory
	output = command.PouchRun("inspect", "-f", "{{.HostConfig.Memory}}", name).Stdout()
	c.Assert(output, check.Equals, fmt.Sprintf("%d\n", result[0].HostConfig.Memory))
}

// TestInspectWrongFormat is to verify using wrong format flag of inspect command.
func (suite *PouchInspectSuite) TestInspectWrongFormat(c *check.C) {
	name := "inspect-wrong-format-print"

	res := command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("inspect", "-f", "{{.NotExists}}", name)
	c.Assert(res.Stderr(), check.NotNil)

	expectString := "Template parsing error"
	if out := res.Combined(); !strings.Contains(out, expectString) {
		c.Fatalf("unexpected output %s expected %s", out, expectString)
	}
}

// TestMultiInspect is to verify inspect command with multiple args.
func (suite *PouchInspectSuite) TestMultiInspect(c *check.C) {
	names := []string{
		"multi-inspect-print-1",
		"multi-inspect-print-2",
	}
	setUp := func() {
		for _, name := range names {
			command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage).Assert(c, icmd.Success)
		}
	}
	cleanUp := func() {
		for _, name := range names {
			DelContainerForceMultyTime(c, name)
		}
	}
	setUp()
	defer cleanUp()

	output := command.PouchRun("inspect", "-f", "{{.Name}}", names[0], names[1]).Stdout()
	expectedOutput := "multi-inspect-print-1\nmulti-inspect-print-2\n"
	c.Assert(output, check.Equals, expectedOutput)
}

// TestMultiInspect is to verify inspect command with multiple args.
func (suite *PouchInspectSuite) TestMultiInspectErrors(c *check.C) {

	errorCases := []struct {
		containers     []string
		args           []string
		expectedOutput string
	}{
		{
			containers: []string{},
			args:       []string{"multi-inspect-print-1", "multi-inspect-print-2"},
			expectedOutput: "\nError: Fetch object error: {\"message\":\"container: multi-inspect-print-1: not found\"}\n" +
				"Error: Fetch object error: {\"message\":\"container: multi-inspect-print-2: not found\"}\n",
		},
		{
			containers: []string{"multi-inspect-print-1"},
			args:       []string{"multi-inspect-print-1", "multi-inspect-print-2"},
			expectedOutput: "multi-inspect-print-1\n" +
				"Error: Fetch object error: {\"message\":\"container: multi-inspect-print-2: not found\"}\n",
		},
	}

	runContainers := func(names []string) {
		for _, name := range names {
			command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage).Assert(c, icmd.Success)
		}
	}
	delContainers := func(names []string) {
		for _, name := range names {
			DelContainerForceMultyTime(c, name)
		}
	}

	for _, errCase := range errorCases {
		runContainers(errCase.containers)
		defer delContainers(errCase.containers)
		res := command.PouchRun("inspect", "-f", "{{.Name}}", errCase.args[0], errCase.args[1])
		c.Assert(res.Stderr(), check.NotNil)
		output := res.Combined()
		c.Assert(output, check.Equals, errCase.expectedOutput)
	}
}
