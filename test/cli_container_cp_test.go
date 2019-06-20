package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchContainerCopySuite is the test suite for container cp CLI.
type PouchContainerCopySuite struct{}

func init() {
	check.Suite(&PouchContainerCopySuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchContainerCopySuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// Test pouch cp, basic usage
func (suite *PouchContainerCopySuite) TestPouchCopy(c *check.C) {
	testDataPath, err := ioutil.TempDir("", "test-pouch-copy")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(testDataPath)

	// test copy from container
	name := "TestPouchCopy"
	command.PouchRun("run",
		"--name", name,
		"-d", busyboxImage,
		"sh", "-c",
		"echo 'test pouch cp' >> data.txt && sleep 10000").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	localTestPath := fmt.Sprintf("%s/%s", testDataPath, "data.txt")
	containerTestPath := fmt.Sprintf("%s:%s", name, "test.txt")

	command.PouchRun("cp", fmt.Sprintf("%s:%s", name, "data.txt"), localTestPath).Assert(c, icmd.Success)
	checkFileContains(c, localTestPath, "test pouch cp")

	// test copy to container
	command.PouchRun("cp", localTestPath, containerTestPath).Assert(c, icmd.Success)
	res := command.PouchRun("exec", name, "cat", "test.txt")
	res.Assert(c, icmd.Success)
	err = util.PartialEqual(res.Stdout(), "test pouch cp")
	c.Assert(err, check.IsNil)

	// test copy to container with non-dir
	res = command.PouchRun("cp", testDataPath, containerTestPath)
	err = util.PartialEqual(res.Combined(), "Error: cannot copy directory")
	c.Assert(err, check.IsNil)
}

// Test pouch cp, where dir locate in volume
func (suite *PouchContainerCopySuite) TestVolumeCopy(c *check.C) {
	testDataPath, err := ioutil.TempDir("", "test-volume-copy")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(testDataPath)

	// test mount rw and copy from container
	name := "TestVolumeCopy"
	command.PouchRun("volume", "create", "--name", name).Assert(c, icmd.Success)
	defer command.PouchRun("volume", "remove", name)
	command.PouchRun("run",
		"--name", name,
		"-d",
		"-v", fmt.Sprintf("%s:%s:rw", name, "/test"),
		busyboxImage,
		"sh", "-c",
		"echo 'test pouch cp' >> /test/data.txt && sleep 10000").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	localTestPath := fmt.Sprintf("%s/%s", testDataPath, "data.txt")
	containerTestPath := fmt.Sprintf("%s:%s", name, "/test/test.txt")

	command.PouchRun("cp", fmt.Sprintf("%s:%s", name, "/test/data.txt"), localTestPath).Assert(c, icmd.Success)
	checkFileContains(c, localTestPath, "test pouch cp")

	// test mount rw and copy to container
	command.PouchRun("cp", localTestPath, containerTestPath).Assert(c, icmd.Success)
	res := command.PouchRun("exec", name, "cat", "/test/test.txt")
	res.Assert(c, icmd.Success)
	err = util.PartialEqual(res.Stdout(), "test pouch cp")
	c.Assert(err, check.IsNil)

	// test mount only ro
	nameRO := "TestVolumeCopyRO"
	command.PouchRun("volume", "create", "--name", nameRO).Assert(c, icmd.Success)
	defer command.PouchRun("volume", "remove", nameRO)
	command.PouchRun("run",
		"--name", nameRO,
		"-d",
		"-v", fmt.Sprintf("%s:%s:ro", nameRO, "/test"),
		busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, nameRO)

	command.PouchRun("cp", localTestPath, containerTestPath).Assert(c, icmd.Success)
	err = util.PartialEqual(res.Stdout(), "can't extract to dir because volume read only")
	c.Assert(err, check.NotNil)
}

// TestStopContainerCopy tests stopped container can work well
func (suite *PouchContainerCopySuite) TestStopContainerCopy(c *check.C) {
	testDataPath := "testdata/cp/test-stop-container-copy"
	c.Assert(os.MkdirAll(testDataPath, 0755), check.IsNil)
	defer os.RemoveAll(testDataPath)

	name := "TestStopContainerCopy"
	command.PouchRun("run", "-d",
		"--name", name,
		busyboxImage,
		"sh", "-c",
		"echo 'test pouch cp' >> data.txt && top").Assert(c, icmd.Success)
	command.PouchRun("stop", "-t", "1", name).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// test copy from container
	localTestPath := fmt.Sprintf("%s/%s", testDataPath, "TestStopContainerCopy.txt")
	containerTestPath := fmt.Sprintf("%s:%s", name, "data.txt")
	command.PouchRun("cp", containerTestPath, localTestPath).Assert(c, icmd.Success)
	checkFileContains(c, localTestPath, "test pouch cp")

	// test stopped container can start after cp
	command.PouchRun("start", name).Assert(c, icmd.Success)
}

// Test pouch cp, where path contains dot
func (suite *PouchContainerCopySuite) TestCopyPathDot(c *check.C) {
	testDataPath := "testdata/cp/test-copy-path-dot"
	c.Assert(os.MkdirAll(testDataPath, 0755), check.IsNil)
	defer os.RemoveAll(testDataPath)

	name := "TestCopyPathDot"
	command.PouchRun("run", "-d",
		"--name", name,
		busyboxImage,
		"sh", "-c",
		"mkdir -p test && echo 'test pouch cp' >> test/data.txt && top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// don't copy test dir under testDataPath
	localTestPath := fmt.Sprintf("%s/%s", testDataPath, "data.txt")
	containerTestPath := fmt.Sprintf("%s:%s", name, "test/.")
	command.PouchRun("cp", containerTestPath, testDataPath).Assert(c, icmd.Success)
	checkFileContains(c, localTestPath, "test pouch cp")
}
