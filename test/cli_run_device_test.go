package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunDeviceSuite is the test suite for run CLI.
type PouchRunDeviceSuite struct{}

func init() {
	check.Suite(&PouchRunDeviceSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunDeviceSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	environment.PruneAllContainers(apiClient)
	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunDeviceSuite) TearDownTest(c *check.C) {
}

// TestRunDeviceMapping is to verify --device param when running a container.
func (suite *PouchRunDeviceSuite) TestRunDeviceMapping(c *check.C) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		c.Skip("Host does not have /dev/zero")
	}

	name := "TestRunDeviceMapping"
	testDev := "/dev/testDev"

	res := command.PouchRun("run",
		"--name", name,
		"--device", fmt.Sprintf("/dev/zero:%s", testDev),
		busyboxImage,
		"ls", testDev)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, testDev) {
		c.Fatalf("unexpected output %s expected %s\n", out, testDev)
	}
}

// TestRunDevicePermissions is to verify --device permissions mode
// when running a container.
func (suite *PouchRunDeviceSuite) TestRunDevicePermissions(c *check.C) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		c.Skip("Host does not have /dev/zero")
	}

	name := "TestRunDevicePermissions"
	testDev := "/dev/testDev"
	permissions := "crw-rw-rw-"

	res := command.PouchRun("run",
		"--name", name,
		"--device", fmt.Sprintf("/dev/zero:%s:rwm", testDev),
		busyboxImage,
		"ls", "-l", testDev)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.HasPrefix(out, permissions) {
		c.Fatalf("Output should begin with %s, got %s\n", permissions, out)
	}
}

// TestRunDeviceInvalidMode is to verify --device wrong mode
// when running a container.
func (suite *PouchRunDeviceSuite) TestRunDeviceInvalidMode(c *check.C) {
	name := "TestRunDeviceInvalidMode"
	wrongMode := "rxm"

	res := command.PouchRun("run",
		"--name", name,
		"--device", fmt.Sprintf("/dev/zero:/dev/zero:%s"+wrongMode),
		busyboxImage,
		"ls", "/dev/zero")
	defer DelContainerForceMultyTime(c, name)

	c.Assert(res.Stderr(), check.NotNil)

	expected := "invalid device mode"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s unexpected output %s. \n",
			expected, out)
	}
}

// TestRunDeviceDirectory is to verify --device
// with a device directory when running a container.
func (suite *PouchRunDeviceSuite) TestRunDeviceDirectory(c *check.C) {
	if _, err := os.Stat("/dev/snd"); err != nil {
		c.Skip("Host does not have direcory /dev/snd")
	}

	name := "TestRunDeviceDirectory"
	srcDev := "/dev/snd"

	res := command.PouchRun("run",
		"--name", name,
		"--device", fmt.Sprintf("%s:/dev:rwm", srcDev),
		busyboxImage,
		"ls", "-l", "/dev")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	// /dev/snd contans two device: timer, seq
	expected := "timer"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s, got %s\n", expected, out)
	}
}

// TestRunWithBadDevice is to verify --device
// with bad device dir when running a container.
func (suite *PouchRunDeviceSuite) TestRunDeviceWithBadDevice(c *check.C) {
	name := "TestRunDeviceWithBadDevice"

	res := command.PouchRun("run",
		"--name", name,
		"--device", "/etc",
		busyboxImage,
		"ls", "/etc")
	defer DelContainerForceMultyTime(c, name)

	c.Assert(res.Stderr(), check.NotNil)

	expected := "not a device node"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s unexpected output %s. \n",
			expected, out)
	}
}

// TestRunDeviceReadWriteBpsIops tests running container
// with a device combined with flags
// --device-read-bps,
// --device-write-bps,
// --device-read-iops,
// --device-write-iops.
func (suite *PouchRunDeviceSuite) TestRunDeviceReadWriteBpsIops(c *check.C) {
	cname := "TestRunDeviceReadWriteBpsIops"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	res := command.PouchRun("run", "-d",
		"--name", cname,
		"--device-read-bps", fmt.Sprintf("%s:1mb", testDisk),
		"--device-write-bps", fmt.Sprintf("%s:1mb", testDisk),
		"--device-read-iops", fmt.Sprintf("%s:1000", testDisk),
		"--device-write-iops", fmt.Sprintf("%s:1000", testDisk),
		busyboxImage,
		"sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioDeviceReadBps), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceReadBps[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceReadBps[0].Rate, check.Equals, uint64(1048576))

	c.Assert(len(result[0].HostConfig.BlkioDeviceWriteBps), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteBps[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteBps[0].Rate, check.Equals, uint64(1048576))

	c.Assert(len(result[0].HostConfig.BlkioDeviceReadIOps), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceReadIOps[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceReadIOps[0].Rate, check.Equals, uint64(1000))

	c.Assert(len(result[0].HostConfig.BlkioDeviceWriteIOps), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteIOps[0].Path, check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteIOps[0].Rate, check.Equals, uint64(1000))

	// test if cgroup has record the real value
	containerID := result[0].ID
	commonDir := "/sys/fs/cgroup/blkio/default"
	path := fmt.Sprintf("%s/%s/blkio.throttle.read_bps_device", commonDir, containerID)
	checkFileContains(c, path, "1048576")
	path = fmt.Sprintf("%s/%s/blkio.throttle.write_bps_device", commonDir, containerID)
	checkFileContains(c, path, "1048576")
	path = fmt.Sprintf("%s/%s/blkio.throttle.read_iops_device", commonDir, containerID)
	checkFileContains(c, path, "1000")
	path = fmt.Sprintf("%s/%s/blkio.throttle.write_iops_device", commonDir, containerID)
	checkFileContains(c, path, "1000")
}
