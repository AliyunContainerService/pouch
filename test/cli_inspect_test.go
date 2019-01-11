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

// TestInspectCreateAndStartedFormat is to verify the format flag of inspect command.
func (suite *PouchInspectSuite) TestInspectCreateAndStartedFormat(c *check.C) {
	name := "TestInspectCreateAndStartedFormat"
	// get root dir
	rootDir, err := GetRootDir()
	if err != nil || rootDir == "" {
		c.Fatalf("failed to get daemon root dir, err(%v)", err)
	}

	// create a raw container
	res := command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage, "top")
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

	// Inspect LogPath, LogPath should be empty before container's start
	output = command.PouchRun("inspect", "-f", "{{.LogPath}}", name).Stdout()
	c.Assert(strings.TrimSpace(output), check.Equals, "")

	// start the created container
	res = command.PouchRun("start", name)
	res.Assert(c, icmd.Success)

	// Inspect LogPath, HostnamePath, HostsPath, ResolvConfPath, Privileged
	output = command.PouchRun("inspect", "-f", "{{.LogPath}}", name).Stdout()
	expectedLogPath := fmt.Sprintf(rootDir+"/containers/%s/json.log", containerID)
	c.Assert(strings.TrimSpace(output), check.Equals, expectedLogPath)

	output = command.PouchRun("inspect", "-f", "{{.ResolvConfPath}}", name).Stdout()
	expectedLogPath = fmt.Sprintf(rootDir+"/containers/%s/resolv.conf", containerID)
	c.Assert(strings.TrimSpace(output), check.Equals, expectedLogPath)

	output = command.PouchRun("inspect", "-f", "{{.HostnamePath}}", name).Stdout()
	expectedLogPath = fmt.Sprintf(rootDir+"/containers/%s/hostname", containerID)
	c.Assert(strings.TrimSpace(output), check.Equals, expectedLogPath)

	output = command.PouchRun("inspect", "-f", "{{.HostsPath}}", name).Stdout()
	expectedLogPath = fmt.Sprintf(rootDir+"/containers/%s/hosts", containerID)
	c.Assert(strings.TrimSpace(output), check.Equals, expectedLogPath)

	output = command.PouchRun("inspect", "-f", "{{.HostConfig.Privileged}}", name).Stdout()
	c.Assert(strings.TrimSpace(output), check.Equals, "false")
}

// TestInspectWrongFormat is to verify using wrong format flag of inspect command.
func (suite *PouchInspectSuite) TestInspectWrongFormat(c *check.C) {
	name := "inspect-wrong-format-print"

	res := command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage, "top")
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
			command.PouchRun("create", "-m", "30M", "--name", name, busyboxImage, "top").Assert(c, icmd.Success)
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
			expectedOutput: "\nError: Fetch object error: {\"message\":\"container multi-inspect-print-1: not found\"}\n" +
				"Error: Fetch object error: {\"message\":\"container multi-inspect-print-2: not found\"}\n",
		},
		{
			containers: []string{"multi-inspect-print-1"},
			args:       []string{"multi-inspect-print-1", "multi-inspect-print-2"},
			expectedOutput: "multi-inspect-print-1\n" +
				"Error: Fetch object error: {\"message\":\"container multi-inspect-print-2: not found\"}\n",
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

//
func (suite *PouchInspectSuite) TestContainerInspectState(c *check.C) {
	// 1. /bin/sh will exit immediately
	name := "TestContainerInspectStateAutoExit"
	res := command.PouchRun("run", "--name", name, busyboxImage, "/bin/sh")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", "-f", "{{.State.Pid}}", name).Stdout()
	c.Assert(strings.TrimSpace(output), check.Equals, "0")
	output = command.PouchRun("inspect", "-f", "{{.State.ExitCode}}", name).Stdout()
	c.Assert(strings.TrimSpace(output), check.Equals, "0")
	output = command.PouchRun("inspect", "-f", "{{.State.Status}}", name).Stdout()
	c.Assert(strings.TrimSpace(output), check.Equals, "exited")

	// 2. stop a container and check the exit code and
	name = "TestContainerInspectStateManullyStop"
	res = command.PouchRun("run", "-d", "--name", name, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)
	// stop container
	res = command.PouchRun("stop", "-t", "1", name)
	res.Assert(c, icmd.Success)

	output = command.PouchRun("inspect", "-f", "{{.State.Pid}}", name).Stdout()
	c.Assert(strings.TrimSpace(output), check.Equals, "0")
	output = command.PouchRun("inspect", "-f", "{{.State.ExitCode}}", name).Stdout()
	c.Assert(strings.TrimSpace(output), check.Not(check.Equals), "0")
	output = command.PouchRun("inspect", "-f", "{{.State.Status}}", name).Stdout()
	c.Assert(strings.TrimSpace(output), check.Equals, "stopped")
}

func (suite *PouchInspectSuite) TestContainerInspectPorts(c *check.C) {
	name := "TestContainerInspectPorts"
	command.PouchRun("run", "-d", "--name", name, "-p", "8080:80", busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()

	containers := make([]types.ContainerJSON, 1)
	err := json.Unmarshal([]byte(output), &containers)
	if err != nil || len(containers) == 0 {
		c.Fatal("fail to format container json")
	}
	data, _ := json.Marshal(containers[0].NetworkSettings.Ports)
	c.Assert(string(data), check.Equals, "{\"80/tcp\":[{\"HostIp\":\"0.0.0.0\",\"HostPort\":\"8080\"}]}")
}
