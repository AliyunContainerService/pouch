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

// PouchCreateSuite is the test suite fo help CLI.
type PouchCreateSuite struct{}

func init() {
	check.Suite(&PouchCreateSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchCreateSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchCreateSuite) TearDownTest(c *check.C) {
}

// TestCreateName is to verify the correctness of creating contaier with specified name.
func (suite *PouchCreateSuite) TestCreateName(c *check.C) {
	name := "create-normal"
	res := command.PouchRun("create", "--name", name, busyboxImage)

	res.Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output %s expected %s\n", out, name)
	}
}

// TestCreateDuplicateContainerName is to verify duplicate container names.
func (suite *PouchCreateSuite) TestCreateDuplicateContainerName(c *check.C) {
	name := "duplicate"

	res := command.PouchRun("create", "--name", name, busyboxImage)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("create", "--name", name, busyboxImage)
	c.Assert(res.Error, check.NotNil)

	if out := res.Combined(); !strings.Contains(out, "already exist") {
		c.Fatalf("unexpected output %s expected already exist\n", out)
	}
}

// TestCreateWithArgs is to verify args.
//
// TODO: pouch inspect should return args info
func (suite *PouchCreateSuite) TestCreateWithArgs(c *check.C) {
	res := command.PouchRun("create", busyboxImage, "/bin/ls")
	res.Assert(c, icmd.Success)
}

// TestCreateWithTTY is to verify tty flag.
//
// TODO: pouch inspect should return tty info
func (suite *PouchCreateSuite) TestCreateWithTTY(c *check.C) {
	res := command.PouchRun("create", "-t", busyboxImage)
	res.Assert(c, icmd.Success)
}

// TestPouchCreateVolume is to verify volume flag.
//
// TODO: pouch inspect should return volume info to check
func (suite *PouchCreateSuite) TestPouchCreateVolume(c *check.C) {
	res := command.PouchRun("create", "-v /tmp:/tmp", busyboxImage)
	res.Assert(c, icmd.Success)
}

// TestCreateInWrongWay tries to run create in wrong way.
func (suite *PouchCreateSuite) TestCreateInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "unknown flag", args: "-a"},

		// TODO: should add the following cases if ready
		// {name: "missing image name", args: ""},
	} {
		res := command.PouchRun("create", tc.args)
		c.Assert(res.Error, check.NotNil, check.Commentf(tc.name))
	}
}

// TestCreateWithLabels tries to test create a container with label.
func (suite *PouchCreateSuite) TestCreateWithLabels(c *check.C) {
	label := "abc=123"
	name := "create-label"

	res := command.PouchRun("create", "--name", name, "-l", label, busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.Config.Labels, check.NotNil)

	if result.Config.Labels["abc"] != "123" {
		c.Errorf("failed to set label: %s", label)
	}
}

// TestCreateEnableLxcfs tries to test create a container with lxcfs.
func (suite *PouchCreateSuite) TestCreateEnableLxcfs(c *check.C) {
	name := "create-lxcfs"

	res := command.PouchRun("create", "--name", name, "--enableLxcfs=true", busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.HostConfig.EnableLxcfs, check.NotNil)

	if result.HostConfig.EnableLxcfs != true {
		c.Errorf("failed to set EnableLxcfs")
	}
}

// TestCreateWithEnv tests creating container with env
func (suite *PouchCreateSuite) TestCreateWithEnv(c *check.C) {
	name := "TestCreateWithEnv"

	res := command.PouchRun("create", "--name", name, "-e TEST=true", busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	ok := false
	for _, v := range result.Config.Env {
		if strings.Contains(v, "TEST=true") {
			ok = true
		}
	}
	c.Assert(ok, check.Equals, true)
}

// TestCreateWithWorkDir tests creating container with a workdir works.
func (suite *PouchCreateSuite) TestCreateWithWorkDir(c *check.C) {
	name := "TestCreateWithWorkDir"

	res := command.PouchRun("create", "--name", name, "-w /tmp/test", busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(strings.TrimSpace(result.Config.WorkingDir), check.Equals, "/tmp/test")

	// TODO: check the work directory has been created.
}
