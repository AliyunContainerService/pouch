package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunBlkioSuite is the test suite for run CLI.
type PouchRunBlkioSuite struct{}

func init() {
	check.Suite(&PouchRunBlkioSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunBlkioSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunBlkioSuite) TearDownTest(c *check.C) {
}

// TestRunBlockIOWeight tests running container with --blkio-weight flag.
func (suite *PouchRunBlkioSuite) TestRunBlockIOWeight(c *check.C) {
	cname := "TestRunBlockIOWeight"
	strvalue := "100"

	res := command.PouchRun("run", "-d", "--blkio-weight", strvalue,
		"--name", cname, busyboxImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	// test if the value is in inspect result
	blkioWeight, err := inspectFilter(cname, ".HostConfig.BlkioWeight")
	c.Assert(err, check.IsNil)
	c.Assert(blkioWeight, check.Equals, strvalue)

	// test if cgroup has record the real value
	containerID, err := inspectFilter(cname, ".ID")
	c.Assert(err, check.IsNil)
	path := fmt.Sprintf(
		"/sys/fs/cgroup/blkio/default/%s/blkio.weight", containerID)
	checkFileContains(c, path, strvalue)

	// test if the value is correct in container
	blkioWeightFile := "/sys/fs/cgroup/blkio/blkio.weight"
	res = command.PouchRun("exec", cname, "cat", blkioWeightFile)
	res.Assert(c, icmd.Success)

	out := res.Stdout()
	c.Assert(out, check.Equals, strvalue+"\n")
}

// TestRunBlockIOWeightDevice tests running container
// with --blkio-weight-device flag.
func (suite *PouchRunBlkioSuite) TestRunBlockIOWeightDevice(c *check.C) {
	cname := "TestRunBlockIOWeightDevice"
	value := 100
	testDisk, found := environment.FindDisk()
	if !found {
		c.Skip("fail to find available disk for blkio test")
	}

	SkipIfFalse(c, func() bool {
		file := fmt.Sprintf("/sys/block/%s/queue/scheduler",
			strings.TrimPrefix(testDisk, "/dev/"))
		if data, err := ioutil.ReadFile(file); err == nil {
			return strings.Contains(string(data), "[cfq]")
		}
		return false
	})

	res := command.PouchRun("run", "-d",
		"--blkio-weight-device", testDisk+":"+strconv.Itoa(value),
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

	c.Assert(len(result[0].HostConfig.BlkioWeightDevice), check.Equals, 1)
	c.Assert(result[0].HostConfig.BlkioWeightDevice[0].Path,
		check.Equals, testDisk)
	c.Assert(result[0].HostConfig.BlkioWeightDevice[0].Weight,
		check.Equals, uint16(value))

	number, exist := util.GetMajMinNumOfDevice(testDisk)
	if !exist {
		c.Skip("fail to get major:minor device number")
	}

	expected := fmt.Sprintf("%s %d\n", number, value)

	// test if the value is correct on the host
	containerID := result[0].ID
	path := fmt.Sprintf(
		"/sys/fs/cgroup/blkio/default/%s/blkio.weight_device", containerID)
	checkFileContains(c, path, strings.Trim(expected, "\n"))

	// test if the value is correct in container
	blkioWeightDevFile := "/sys/fs/cgroup/blkio/blkio.weight_device"
	res = command.PouchRun("exec", cname, "cat", blkioWeightDevFile)
	res.Assert(c, icmd.Success)

	out := res.Stdout()
	c.Assert(out, check.Equals, expected)
}

// TestRunWithBlkioWeight is to verify --specific Blkio Weight
// when running a container.
func (suite *PouchRunBlkioSuite) TestRunWithBlkioWeight(c *check.C) {
	name := "test-run-with-blkio-weight"

	res := command.PouchRun("run", "-d", "--name", name,
		"--blkio-weight", "500", busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)
}
