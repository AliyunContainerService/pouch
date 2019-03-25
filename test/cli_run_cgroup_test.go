package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

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

	res = command.PouchRun("exec", name, "cat", "/sys/fs/cgroup/memory/memory.limit_in_bytes")
	res.Assert(c, icmd.Success)
	c.Assert(util.PartialEqual(res.Stdout(), "314572800"), check.IsNil)

	res = command.PouchRun("exec", name, "cat", "/proc/self/cgroup")
	res.Assert(c, icmd.Success)
	cgroupPaths := util.ParseCgroupFile(res.Stdout())

	for _, v := range cgroupPaths {
		// NOTE: if container in child cgroup namespace, cgroup mount is /,
		// skip test since we can not get total cgroup path
		if v == "/" {
			break
		}
		if !strings.Contains(v, cgroupParent) {
			c.Fatalf("unexpected cgroup path %v, expect to has %s in path", v, cgroupParent)
		}
	}

	// inspect Container ID
	res = command.PouchRun("inspect", "-f", "{{.ID}}", name)
	res.Assert(c, icmd.Success)
	containerID := strings.TrimSpace(res.Stdout())

	// if cgroupParent is absolute path, cgroup path should /cgroupMount/subsystem/
	if filepath.IsAbs(cgroupParent) {
		for p := range cgroupPaths {
			// like name=systemd, and rdma not created by runc
			if strings.Contains(p, "=") || strings.Contains(p, "rdma") {
				continue
			}
			if _, err := os.Stat(filepath.Join("/sys/fs/cgroup", p, cgroupParent, containerID)); err != nil {
				c.Fatalf("%s cgroup path should exist", filepath.Join("/sys/fs/cgroup", p, cgroupParent, containerID))
			}
		}
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
