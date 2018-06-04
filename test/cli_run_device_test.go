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

	name := "test-run-device-mapping"
	testDev := "/dev/testDev"

	res := command.PouchRun("run", "--name", name,
		"--device", "/dev/zero:"+testDev,
		busyboxImage, "ls", testDev)
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

	name := "test-run-device-permissions"
	testDev := "/dev/testDev"
	permissions := "crw-rw-rw-"

	res := command.PouchRun("run", "--name", name, "--device",
		"/dev/zero:"+testDev+":rwm", busyboxImage, "ls", "-l", testDev)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.HasPrefix(out, permissions) {
		c.Fatalf("Output should begin with %s, got %s\n", permissions, out)
	}
}

// TestRunDeviceInvalidMode is to verify --device wrong mode
// when running a container.
func (suite *PouchRunDeviceSuite) TestRunDeviceInvalidMode(c *check.C) {
	name := "test-run-device-with-wrong-mode"
	wrongMode := "rxm"

	res := command.PouchRun("run", "--name", name, "--device",
		"/dev/zero:/dev/zero:"+wrongMode, busyboxImage, "ls", "/dev/zero")
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

	name := "test-run-with-directory-device"
	srcDev := "/dev/snd"

	res := command.PouchRun("run", "--name", name, "--device",
		srcDev+":/dev:rwm", busyboxImage, "ls", "-l", "/dev")
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
	name := "test-run-with-bad-device"

	res := command.PouchRun("run", "--name", name, "--device", "/etc",
		busyboxImage, "ls", "/etc")
	defer DelContainerForceMultyTime(c, name)

	c.Assert(res.Stderr(), check.NotNil)

	expected := "not a device node"
	if out := res.Combined(); !strings.Contains(out, expected) {
		c.Fatalf("Output should contain %s unexpected output %s. \n",
			expected, out)
	}
}

// TestRunDeviceReadBps tests running container with --device-read-bps flag.
func (suite *PouchRunDeviceSuite) TestRunDeviceReadBps(c *check.C) {
	cname := "TestRunDeviceReadBps"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	res := command.PouchRun("run", "-d",
		"--device-read-bps", testDisk+":1mb",
		"--name", cname, busyboxImage, "sleep", "10000")
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
	c.Assert(result[0].HostConfig.BlkioDeviceReadBps[0].Path,
		check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceReadBps[0].Rate,
		check.Equals, uint64(1048576))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/blkio/default/%s/blkio.throttle.read_bps_device",
			containerID)
		checkFileContains(c, path, "1048576")
	}
}

// TestRunDeviceWriteBps tests running container
// with --device-write-bps flag.
func (suite *PouchRunDeviceSuite) TestRunDeviceWriteBps(c *check.C) {
	cname := "TestRunDeviceWriteBps"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	res := command.PouchRun("run", "-d",
		"--device-write-bps", testDisk+":1mb",
		"--name", cname, busyboxImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioDeviceWriteBps),
		check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteBps[0].Path,
		check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteBps[0].Rate,
		check.Equals, uint64(1048576))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/blkio/default/%s/blkio.throttle.write_bps_device",
			containerID)
		checkFileContains(c, path, "1048576")
	}
}

// TestRunDeviceReadIops tests running container with --device-read-iops flag.
func (suite *PouchRunDeviceSuite) TestRunDeviceReadIops(c *check.C) {
	cname := "TestRunDeviceReadIops"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	res := command.PouchRun("run", "-d",
		"--device-read-iops", testDisk+":1000",
		"--name", cname, busyboxImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioDeviceReadIOps),
		check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceReadIOps[0].Path,
		check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceReadIOps[0].Rate,
		check.Equals, uint64(1000))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/blkio/default/%s/blkio.throttle.read_iops_device",
			containerID)
		checkFileContains(c, path, "1000")
	}
}

// TestRunDeviceWriteIops tests running container
// with --device-write-iops flag.
func (suite *PouchRunDeviceSuite) TestRunDeviceWriteIops(c *check.C) {
	cname := "TestRunDeviceWriteIops"
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	res := command.PouchRun("run", "-d",
		"--device-write-iops", testDisk+":1000",
		"--name", cname, busyboxImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	res = command.PouchRun("inspect", cname)
	res.Assert(c, icmd.Success)

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(res.Stdout()), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(result[0].HostConfig.BlkioDeviceWriteIOps),
		check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteIOps[0].Path,
		check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioDeviceWriteIOps[0].Rate,
		check.Equals, uint64(1000))

	// test if cgroup has record the real value
	containerID := result[0].ID
	{
		path := fmt.Sprintf(
			"/sys/fs/cgroup/blkio/default/%s/blkio.throttle.write_iops_device",
			containerID)
		checkFileContains(c, path, "1000")
	}
}
