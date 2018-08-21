package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/pkg/errors"
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

// TestRunWithVolumesDestinationNotEmpty tests running container with a volume whose
// destination path in the image is not empty, we should populate the volume with
// corresponding content in the image.
func (suite *PouchRunVolumeSuite) TestRunWithVolumesDestinationNotEmpty(c *check.C) {
	// The image `calico/cni:v3.1.3` has an anonymous volume whose destination path
	// is `/opt/cni` which is not empty in the image. We should still be able to see
	// the data in the image's `/opt/cni` when the volume is mounted.
	// Details refer to: https://github.com/alibaba/pouch/issues/1739
	image := "calico/cni:v3.1.3"
	containerName := "volumesDestinationNotEmpty"

	// For the workdir of image `calico/cni:v3.1.3` is `/opt/cni/bin`,
	// it is enough to prove this feature works fine if we could run
	// the container with default workdir successfully.
	res := command.PouchRun("run", "--name", containerName, image, "ls")
	defer DelContainerForceMultyTime(c, containerName)
	res.Assert(c, icmd.Success)
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

// TestRunWithVolumesFromDifferentSources tests containers with volumes from different sources
func (suite *PouchRunVolumeSuite) TestRunWithVolumesFromDifferentSources(c *check.C) {
	// TODO: build the image with volume
	imageWithVolume := "registry.hub.docker.com/shaloulcy/busybox:with-volume"
	containerName1 := "TestRunWithVolumesFromImage"
	containerName2 := "TestRunWithVolumesFromContainerAndImage"

	// pull the image with volume
	command.PouchRun("pull", imageWithVolume).Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("rmi", "-f", imageWithVolume).Assert(c, icmd.Success)
	}()

	// start the container1 which has volume from image,
	// and the volume destination is /data
	command.PouchRun("run", "-d", "-t",
		"--name", containerName1, imageWithVolume, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerName1)

	out1 := command.PouchRun("inspect", containerName1).Assert(c, icmd.Success)
	container1 := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(out1.Stdout()), &container1); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(len(container1[0].Mounts), check.Equals, 1)
	volumeName1 := container1[0].Mounts[0].Name

	// start the container2 which has volumes from container1,
	// and has an extra volume whose desctination is /data
	command.PouchRun("run", "-d", "-t",
		"-v", "/data", "--volumes-from", containerName1,
		"--name", containerName2, imageWithVolume, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerName2)

	out2 := command.PouchRun("inspect", containerName2).Assert(c, icmd.Success)
	container2 := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(out2.Stdout()), &container2); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	// only one Mount, as the volumes-from will overwrite the binds
	c.Assert(len(container2[0].Mounts), check.Equals, 1)
	volumeName2 := container2[0].Mounts[0].Name

	// assert the two volumes is the same one
	c.Assert(volumeName1, check.Equals, volumeName2)
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

// TestRunWithDiskQuotaForLog tests limit the size of container's log.
func (suite *PouchRunVolumeSuite) TestRunWithDiskQuotaForLog(c *check.C) {
	if !environment.IsDiskQuota() {
		c.Skip("Host does not support disk quota")
	}

	cname := "TestRunWithDiskQuotaForLog"
	command.PouchRun("run", "-d", "--disk-quota", "10m",
		"--name", cname, busyboxImage, "top").Assert(c, icmd.Success)

	defer DelContainerForceMultyTime(c, cname)

	containerMetaDir, err := getContainerMetaDir(cname)
	c.Assert(err, check.Equals, nil)

	testFile := filepath.Join(containerMetaDir, "diskquota_testfile")
	expct := icmd.Expected{
		ExitCode: 1,
		Err:      "Disk quota exceeded",
	}
	err = icmd.RunCommand("dd", "if=/dev/zero", "of="+testFile, "bs=1M", "count=20", "oflag=direct").Compare(expct)
	c.Assert(err, check.IsNil)
}

func getContainerMetaDir(name string) (string, error) {
	ret := command.PouchRun("inspect", name)

	mergedDir := ""
	for _, line := range strings.Split(ret.Stdout(), "\n") {
		if strings.Contains(line, "MergedDir") {
			mergedDir = strings.Split(line, "\"")[3]
			break
		}
	}

	var (
		graph string
		cid   string
	)
	if mergedDir == "" {
		return "", errors.Errorf("failed to get container metadata directory")
	}

	parts := strings.Split(mergedDir, "/")
	for i, part := range parts {
		if part == "containerd" {
			graph = "/" + filepath.Join(parts[:i]...)
			cid = parts[i+4]
			break
		}
	}

	return filepath.Join(graph, "containers", cid), nil
}
