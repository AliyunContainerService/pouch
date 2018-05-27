package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunVolumeSuite is the test suite for run CLI.
type PouchRunVolumeSuite struct{}

func init() {
	check.Suite(&PouchRunVolumeSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunVolumeSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunVolumeSuite) TearDownTest(c *check.C) {
}

// TestRunWithLocalVolume is to verify run container with -v volume works.
func (suite *PouchRunVolumeSuite) TestRunWithLocalVolume(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	name := funcname
	{
		res := command.PouchRun("volume", "create", "--name", funcname)
		defer func() {
			command.PouchRun("volume", "remove", funcname).Assert(c, icmd.Success)
		}()
		res.Assert(c, icmd.Success)
	}

	{
		res := command.PouchRun("run", "--name", name, "-v", funcname+":/tmp",
			busyboxImage, "touch", "/tmp/test")
		defer DelContainerForceMultyTime(c, name)
		res.Assert(c, icmd.Success)
	}

	// check the existence of /var/lib/pouch/volume/function/test
	icmd.RunCommand("stat",
		DefaultVolumeMountPath+"/"+funcname+"/test").Assert(c, icmd.Success)
}

// TestRunWithHostFileVolume tests binding a host file as a volume into container.
// fixes https://github.com/alibaba/pouch/issues/813
func (suite *PouchRunVolumeSuite) TestRunWithHostFileVolume(c *check.C) {
	// first create a file on the host
	filepath := "/tmp/TestRunWithHostFileVolume.md"
	icmd.RunCommand("touch", filepath).Assert(c, icmd.Success)

	cname := "TestRunWithHostFileVolume"
	res := command.PouchRun("run", "-d", "--name", cname, "-v",
		fmt.Sprintf("%s:%s", filepath, filepath), busyboxImage)

	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)
}

// TestRunWithVolumesFrom tests running container with --volumes-from.
func (suite *PouchRunVolumeSuite) TestRunWithVolumesFrom(c *check.C) {
	volumeName := "volumesfrom-test-volume"
	containerName1 := "volumesfrom-test-1"
	containerName2 := "volumesfrom-test-2"

	// create volume
	command.PouchRun("volume", "create", "-n", volumeName).Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
	}()

	// run container1
	res := command.PouchRun("run", "-d",
		"-v", volumeName+":/mnt",
		"-v", "/tmp:/tmp",
		"--name", containerName1, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, containerName1)
	res.Assert(c, icmd.Success)

	// stop container1
	command.PouchRun("stop", containerName1).Assert(c, icmd.Success)

	// run container2
	res = command.PouchRun("run", "-d",
		"--volumes-from", containerName1,
		"--name", containerName2, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, containerName2)
	res.Assert(c, icmd.Success)

	// inspect container2
	ret := command.PouchRun("inspect", containerName2)
	ret.Assert(c, icmd.Success)
	out := ret.Stdout()

	volumeFound := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "\"volumesfrom-test-volume\"") {
			volumeFound = true
			break
		}
	}

	c.Assert(volumeFound, check.Equals, true)
}

// TestRunWithVolumesFromWithDupclicate tests running container with --volumes-from.
func (suite *PouchRunVolumeSuite) TestRunWithVolumesFromWithDupclicate(c *check.C) {
	volumeName := "volumesfromDupclicate-test-volume"
	containerName1 := "volumesfromDupclicate-test-1"
	containerName2 := "volumesfromDupclicate-test-2"

	// create volume
	command.PouchRun("volume", "create", "-n", volumeName).Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
	}()

	// run container1
	res := command.PouchRun("run", "-d",
		"-v", volumeName+":/mnt",
		"-v", "/tmp:/tmp",
		"--name", containerName1, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, containerName1)
	res.Assert(c, icmd.Success)

	// stop container1
	command.PouchRun("stop", containerName1).Assert(c, icmd.Success)

	// run container2
	res = command.PouchRun("run", "-d",
		"-v", "/tmp:/tmp",
		"--volumes-from", containerName1,
		"--name", containerName2, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, containerName2)
	res.Assert(c, icmd.Success)

	// inspect container2
	ret := command.PouchRun("inspect", containerName2)
	ret.Assert(c, icmd.Success)
	out := ret.Stdout()

	volumeFound := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "\"volumesfromDupclicate-test-volume\"") {
			volumeFound = true
			break
		}
	}

	c.Assert(volumeFound, check.Equals, true)
}

// TestRunWithDiskQuotaRegular tests running container with --disk-quota.
func (suite *PouchRunVolumeSuite) TestRunWithDiskQuotaRegular(c *check.C) {
	if !environment.IsDiskQuota() {
		c.Skip("Host does not support disk quota")
	}

	volumeName := "diskquota-volume"
	containerName := "diskquota-regular"

	ret := command.PouchRun("volume", "create", "-n", volumeName,
		"-o", "opt.size=256m", "-o", "mount=/data/volume")
	defer func() {
		command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
	}()
	ret.Assert(c, icmd.Success)

	ret = command.PouchRun("run",
		"--disk-quota=1024m",
		`--disk-quota=".*=512m"`,
		`--disk-quota="/mnt/mount1=768m"`,
		"-v", "/data/mount1:/mnt/mount1",
		"-v", "/data/mount2:/mnt/mount2",
		"-v", "diskquota-volume:/mnt/mount3",
		"--name", containerName, busyboxImage, "df")
	defer DelContainerForceMultyTime(c, containerName)
	ret.Assert(c, icmd.Success)

	out := ret.Stdout()

	rootFound := false
	mount1Found := false
	mount2Found := false
	mount3Found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "/") &&
			strings.Contains(line, "1048576") {
			rootFound = true
			continue
		}

		if strings.Contains(line, "/mnt/mount1") &&
			strings.Contains(line, "786432") {
			mount1Found = true
			continue
		}

		if strings.Contains(line, "/mnt/mount2") &&
			strings.Contains(line, "524288") {
			mount2Found = true
			continue
		}

		if strings.Contains(line, "/mnt/mount3") &&
			strings.Contains(line, "262144") {
			mount3Found = true
			continue
		}
	}

	c.Assert(rootFound, check.Equals, true)
	c.Assert(mount1Found, check.Equals, true)
	c.Assert(mount2Found, check.Equals, true)
	c.Assert(mount3Found, check.Equals, true)
}

// TestRunWithDiskQuota tests running container with --disk-quota.
func (suite *PouchRunVolumeSuite) TestRunWithDiskQuota(c *check.C) {
	if !environment.IsDiskQuota() {
		c.Skip("Host does not support disk quota")
	}

	cname := "TestRunWithDiskQuota"
	ret := command.PouchRun("run", "--disk-quota", "2000m",
		"--name", cname, busyboxImage, "df")

	defer DelContainerForceMultyTime(c, cname)
	ret.Assert(c, icmd.Success)

	out := ret.Combined()

	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "/") &&
			strings.Contains(line, "2048000") {
			found = true
			break
		}
	}

	c.Assert(found, check.Equals, true)
}
