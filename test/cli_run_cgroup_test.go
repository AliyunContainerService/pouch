package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunCgroupSuite is the test suite for run CLI.
type PouchRunCgroupSuite struct{}

func init() {
	check.Suite(&PouchRunCgroupSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunCgroupSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunCgroupSuite) TearDownTest(c *check.C) {
}

// TestRunWithCgroupParent tests running container with --cgroup-parent.
func (suite *PouchRunCgroupSuite) TestRunWithCgroupParent(c *check.C) {
	// cgroup-parent relative path
	testRunWithCgroupParent(c, "pouch",
		"TestRunWithRelativePathOfCgroupParent")

	// cgroup-parent absolute path
	testRunWithCgroupParent(c, "/pouch/test",
		"TestRunWithAbsolutePathOfCgroupParent")
}

func testRunWithCgroupParent(c *check.C, cgroupParent, name string) {
	res := command.PouchRun("run", "-d", "-m", "300M",
		"--cgroup-parent", cgroupParent,
		"--name", name, busyboxImage, "top")

	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	containerID, err := inspectFilter(name, ".ID")
	c.Assert(err, check.IsNil)

	// this code slice may not robust, but for this test case is enough.
	if strings.HasPrefix(cgroupParent, "/") {
		cgroupParent = cgroupParent[1:]
	}

	if cgroupParent == "" {
		cgroupParent = "default"
	}

	file := "/sys/fs/cgroup/memory/" + cgroupParent + "/" +
		containerID + "/memory.limit_in_bytes"
	if _, err := os.Stat(file); err != nil {
		c.Fatalf("container %s cgroup mountpoint not exists", name)
	}

	out, err := exec.Command("cat", file).Output()
	if err != nil {
		c.Fatalf("execute cat command failed: %v", err)
	}

	if !strings.Contains(string(out), "314572800") {
		c.Fatalf("unexpected output %s expected %s\n",
			string(out), "314572800")
	}

}

// TestRunInvalidCgroupParent checks that a specially-crafted cgroup parent
// doesn't cause pouch to crash or start modifying /.
func (suite *PouchRunCgroupSuite) TestRunInvalidCgroupParent(c *check.C) {
	testRunInvalidCgroupParent(c,
		"../../../../../../../../SHOULD_NOT_EXIST",
		"SHOULD_NOT_EXIST", "cgroup-invalid-test")

	testRunInvalidCgroupParent(c,
		"/../../../../../../../../SHOULD_NOT_EXIST",
		"/SHOULD_NOT_EXIST", "cgroup-absolute-invalid-test")
}

func testRunInvalidCgroupParent(c *check.C, cgroupParent, cleanCgroupParent, name string) {
	command.PouchRun("run", "-m", "300M",
		"--cgroup-parent", cgroupParent, "--name", name, busyboxImage,
		"cat", "/proc/self/cgroup").Assert(c, icmd.Success)

	// We expect "/SHOULD_NOT_EXIST" to not exist.
	// If not, we have a security issue.
	if _, err := os.Stat("/SHOULD_NOT_EXIST"); err == nil ||
		!os.IsNotExist(err) {
		c.Fatalf("SECURITY: --cgroup-parent with " +
			"../../ relative paths cause files to be created" +
			" in the host (this is bad) !!")
	}
}
