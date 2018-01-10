package main

import (
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunSuite is the test suite for help CLI.
type PouchRunSuite struct{}

func init() {
	check.Suite(&PouchRunSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	c.Assert(environment.PruneAllContainers(apiClient), check.IsNil)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunSuite) TearDownTest(c *check.C) {
	c.Assert(environment.PruneAllContainers(apiClient), check.IsNil)
}

// TestRun is to verify the correctness of run container with specified name.
func (suite *PouchRunSuite) TestRun(c *check.C) {
	name := "test-run"

	command.PouchRun("run", "-d", "--name", name, busyboxImage).Assert(c, icmd.Success)

	res := command.PouchRun("ps").Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output %s: should contains container %s\n", out, name)
	}
}

// TestRunPrintHi is to verify run container with executing a command.
func (suite *PouchRunSuite) TestRunPrintHi(c *check.C) {
	name := "test-run-print-hi"

	res := command.PouchRun("run", "--name", name, busyboxImage, "echo", "hi")
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, "hi") {
		c.Fatalf("unexpected output %s expected hi\n", out)
	}
}

// TestRunDeviceMapping is to verify --device param when running a container.
func (suite *PouchRunSuite) TestRunDeviceMapping(c *check.C) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		c.Skip("Host does not have /dev/zero")
	}

	name := "test-run-device-mapping"
	testDev := "/dev/testDev"

	res := command.PouchRun("run", "--name", name, "--device", "/dev/zero:"+testDev, busyboxImage, "ls", testDev)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, testDev) {
		c.Fatalf("unexpected output %s expected %s\n", out, testDev)
	}
}

// TestRunDevicePermissions is to verify --device permissions mode when running a container.
func (suite *PouchRunSuite) TestRunDevicePermissions(c *check.C) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		c.Skip("Host does not have /dev/zero")
	}

	name := "test-run-device-permissions"
	testDev := "/dev/testDev"
	permissions := "crw-rw-rw-"

	res := command.PouchRun("run", "--name", name, "--device", "/dev/zero:"+testDev+":rwm", busyboxImage, "ls", "-l", testDev)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.HasPrefix(out, permissions) {
		c.Fatalf("Output should begin with %s, got %s\n", permissions, out)
	}
}

// TestRunDeviceInvalidMode is to verify --device wrong mode when running a container.
func (suite *PouchRunSuite) TestRunDeviceInvalidMode(c *check.C) {
	name := "test-run-device-with-wrong-mode"
	wrongMode := "rxm"

	res := command.PouchRun("run", "--name", name, "--device", "/dev/zero:/dev/zero:"+wrongMode, busyboxImage, "ls", "/dev/zero")
	c.Assert(res.Error, check.NotNil)

	expected := "invalid device mode"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s unexpected output %s. \n", expected, out)
	}
}

// TestRunDeviceDirectory is to verify --device with a device directory when running a container.
func (suite *PouchRunSuite) TestRunDeviceDirectory(c *check.C) {
	if _, err := os.Stat("/dev/snd"); err != nil {
		c.Skip("Host does not have direcory /dev/snd")
	}

	name := "test-run-with-directory-device"
	srcDev := "/dev/snd"

	res := command.PouchRun("run", "--name", name, "--device", srcDev+":/dev:rwm", busyboxImage, "ls", "-l", "/dev")
	res.Assert(c, icmd.Success)

	// /dev/snd contans two device: timer, seq
	expected := "timer"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s, got %s\n", expected, out)
	}
}

// TestRunWithBadDevice is to verify --device with bad device dir when running a container.
func (suite *PouchRunSuite) TestRunDeviceWithBadDevice(c *check.C) {
	name := "test-run-with-bad-device"

	res := command.PouchRun("run", "--name", name, "--device", "/etc", busyboxImage, "ls", "/etc")
	c.Assert(res.Error, check.NotNil)

	expected := "not a device node"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s unexpected output %s. \n", expected, out)
	}
}

// TestRunInWrongWay tries to run create in wrong way.
func (suite *PouchRunSuite) TestRunInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "unknown flag", args: "-a"},

		// TODO: should add the following cases if ready
		// {name: "missing image name", args: ""},
	} {
		res := command.PouchRun("run", tc.args)
		c.Assert(res.Error, check.NotNil, check.Commentf(tc.name))
	}
}

// TestRunEnableLxcfs is to verify run container with lxcfs.
func (suite *PouchRunSuite) TestRunEnableLxcfs(c *check.C) {
	name := "test-run-lxcfs"

	res := command.PouchRun("run", "--name", name, "--enableLxcfs=true", busyboxImage, "cat", "/proc/uptime")
	res.Assert(c, icmd.Success)

	out := res.Combined()

	// As different test machine may have different uptime, set 10s as the max value.
	maxtime := 10.0
	for _, v := range strings.Fields(out) {
		f, err := strconv.ParseFloat(v, 64)
		c.Assert(err, check.IsNil)

		if math.Max(f, maxtime) == f && math.Abs(f-maxtime) > 0.0001 {
			c.Fatalf("upexpected output %s \n", out)
		}
	}
}
