package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
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
	funcname := "TestRunWithLocalVolume"

	{
		res := command.PouchRun("volume", "create", "--name", funcname)
		defer func() {
			command.PouchRun("volume", "remove", funcname).Assert(c, icmd.Success)
		}()
		res.Assert(c, icmd.Success)
	}

	{
		res := command.PouchRun("run", "--name", funcname, "-v", funcname+":/tmp",
			busyboxImage, "touch", "/tmp/test")
		defer DelContainerForceMultyTime(c, funcname)
		res.Assert(c, icmd.Success)
	}

	// check the existence of /var/lib/pouch/volume/function/test
	icmd.RunCommand("stat",
		DefaultVolumeMountPath+"/"+funcname+"/test").Assert(c, icmd.Success)
}

// TestRunWithTmpFSVolume tests running container with tmpfs volume.
func (suite *PouchRunVolumeSuite) TestRunWithTmpFSVolume(c *check.C) {
	cname := "TestRunWithTmpfsVolume"

	command.PouchRun("volume", "create", "--name", cname, "--driver", "tmpfs",
		"-o", "opt.size=1m").Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("volume", "rm", cname).Assert(c, icmd.Success)
	}()

	res := command.PouchRun("run", "-v", cname+":/opt", "--name", cname,
		busyboxImage, "df", "-h", "/opt")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	c.Assert(strings.Contains(res.Stdout(), "1.0M"), check.Equals, true)
}

//TestRunWithVolumeCopyData tests binds copying data
//Pouch volumes should copy data, but host bind mount should not.
func (suite *PouchRunVolumeSuite) TestRunWithVolumeCopyData(c *check.C) {
	volumeName := "volume-test-copydata"
	hostdir := "/tmp/bind-test-copydata"
	containerName1 := "copydata-test-1"
	containerName2 := "copydata-test-2"

	// create volume
	command.PouchRun("volume", "create", "-n", volumeName).Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
	}()

	// dirs under busybox image `/var` directory
	// notes: there is a `/var/log` directory under rich mode container
	expectedDirs := []string{"spool", "www"}

	command.PouchRun("run", "-t", "-v", volumeName+":/var", "--name", containerName1, busyboxImage, "ls", "/var").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerName1)
	output1 := icmd.RunCommand("ls", DefaultVolumeMountPath+"/"+volumeName).Stdout()

	if !lsResultContains(output1, expectedDirs) {
		c.Fatalf("expected \"%s\" directory under /var directory, but got %s",
			strings.Join(expectedDirs, " "), output1)
	}

	command.PouchRun("run", "-t", "-v", hostdir+":/var", "--name", containerName2, busyboxImage, "ls", "/var").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerName2)
	defer icmd.RunCommand("rm", "-rf", hostdir)
	output2 := icmd.RunCommand("ls", hostdir).Stdout()

	if lsResultContains(output2, expectedDirs) {
		c.Fatalf("volume mount in host bind mode, but \"%s\" exists", strings.Join(expectedDirs, " "))
	}
}

func lsResultContains(res string, names []string) bool {
	lines := strings.Split(res, "\n")
	for _, name := range names {
		name = strings.TrimSpace(name)
		if !utils.StringInSlice(lines, name) {
			return false
		}
	}
	return true
}

// TestRunWithHostFileVolume tests binding a host file as a volume into container.
// fixes https://github.com/alibaba/pouch/issues/813
func (suite *PouchRunVolumeSuite) TestRunWithHostFileVolume(c *check.C) {
	// first create a file on the host
	filepath := "/tmp/TestRunWithHostFileVolume.md"
	icmd.RunCommand("touch", filepath).Assert(c, icmd.Success)
	defer icmd.RunCommand("rm", "-f", filepath)

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
	image := environment.CniRepo + ":" + environment.CniTag
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
	if environment.IsAliKernel() {
		imageWithVolume = "reg.docker.alibaba-inc.com/pouch/busybox:with-volume"
	}
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

func (suite *PouchRunVolumeSuite) TestRunCopyDataWithDR(c *check.C) {
	cname := "TestRunCopyDataWithDR_Container"
	vname := "TestRunCopyDataWithDR_Volume"

	command.PouchRun("volume", "create", "-n", vname).Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", vname)

	command.PouchRun("run", "-d", "--name", cname,
		"-v", vname+":/var/spool:dr",
		"-v", vname+":/var:dr",
		"-v", vname+":/data", busyboxImage, "top").Assert(c, icmd.Success)
	defer command.PouchRun("rm", "-vf", cname)

	res := command.PouchRun("exec", cname, "ls", "/var")
	res.Assert(c, icmd.Success)
	if !strings.Contains(res.Stdout(), "spool") ||
		!strings.Contains(res.Stdout(), "www") {
		c.Fatal("no copy image data, miss spool and www directory")
	}

	res = command.PouchRun("exec", cname, "ls", "/var/spool")
	res.Assert(c, icmd.Success)
	if !strings.Contains(res.Stdout(), "mail") {
		c.Fatal("no copy image data, miss mail directory")
	}
}

func (suite *PouchRunVolumeSuite) TestRunVolumesFromWithDR(c *check.C) {
	vName := "TestRunVolumesFromWithDR_Volume"
	cName := "TestRunVolumesFromWithDR_Container"
	cNameBak := "TestRunVolumesFromWithDR_Container_bak"

	// create volume
	command.PouchRun("volume", "create", "-n", vName).Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", vName)

	// run bak container
	command.PouchRun("run", "-d", "--name", cNameBak,
		"-v", vName+":/var:dr",
		"-v", vName+":/data", busyboxImage, "top").Assert(c, icmd.Success)

	var bakRemoved bool
	defer func() {
		if !bakRemoved {
			command.PouchRun("rm", "-vf", cNameBak)
		}
	}()

	// stop bak container
	command.PouchRun("stop", cNameBak).Assert(c, icmd.Success)

	// run new container with volumes-from
	command.PouchRun("run", "-d", "--name", cName,
		"--volumes-from", cNameBak, busyboxImage, "top").Assert(c, icmd.Success)
	defer command.PouchRun("rm", "-vf", cName)

	// remove bak container
	command.PouchRun("rm", "-vf", cNameBak).Assert(c, icmd.Success)
	bakRemoved = true

	// check inspect mountpoint mode
	ctr, err := apiClient.ContainerGet(context.Background(), cName)
	if err != nil {
		c.Fatalf("failed to get container info, err(%v)", err)
	}

	var found bool
	for _, m := range ctr.Mounts {
		if m.Destination == "/var" && m.Mode == "dr" {
			found = true
		}
	}

	c.Assert(found, check.Equals, true)
}
