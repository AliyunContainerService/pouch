package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchVolumeSuite is the test suite for volume CLI.
type PouchVolumeSuite struct{}

func init() {
	check.Suite(&PouchVolumeSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchVolumeSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllVolumes(apiClient)
	environment.PruneAllContainers(apiClient)
	PullImage(c, busyboxImage)
}

// TestVolumeWorks tests "pouch volume" work.
func (suite *PouchVolumeSuite) TestVolumeWorks(c *check.C) {
	funcname := "TestVolumeWorks"

	command.PouchRun("volume", "create", "--name", funcname).Assert(c, icmd.Success)
	command.PouchRun("volume", "inspect", funcname).Assert(c, icmd.Success)
	command.PouchRun("volume", "remove", funcname).Assert(c, icmd.Success)

	expct := icmd.Expected{
		ExitCode: 1,
		Err:      "No such file or directory",
	}
	err := icmd.RunCommand("stat", DefaultVolumeMountPath+"/"+funcname).Compare(expct)
	c.Assert(err, check.IsNil)
}

// TestVolumeCreateLocalDriverAndSpecifyMountPoint tests "pouch volume create" works.
func (suite *PouchVolumeSuite) TestVolumeCreateLocalDriverAndSpecifyMountPoint(c *check.C) {
	funcname := "TestVolumeCreateLocalDriverAndSpecifyMountPoint"

	res := command.PouchRun("volume", "create", "--name", funcname, "--driver", "local", "-o", "mount=/tmp/"+funcname)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("volume", "inspect", funcname)
	res.Assert(c, icmd.Success)
	output := res.Stdout()
	if !strings.Contains(output, "local") {
		c.Errorf("failed to get the backend driver, expect:local, actually: %s", output)
	}

	if !strings.Contains(output, "/tmp/"+funcname) {
		c.Errorf("failed to get the mountpoint, expect:/tmp, actually: %s", output)
	}

	command.PouchRun("volume", "remove", funcname).Assert(c, icmd.Success)
}

// TestVolumeCreateWithMountPointExitsFile tests when MountPoint is an existing file, returns error.
func (suite *PouchVolumeSuite) TestVolumeCreateWithMountPointExitsFile(c *check.C) {
	funcname := "TestVolumeCreateWithMountPointExitsFile"

	expct := icmd.Expected{
		ExitCode: 1,
		Err:      "mount path is not a dir",
	}

	icmd.RunCommand("touch", "/tmp/"+funcname)

	err := command.PouchRun("volume", "create", "--name", funcname,
		"--driver", "local", "-o", "mount=/tmp/"+funcname).Compare(expct)
	defer command.PouchRun("volume", "remove", funcname)

	c.Assert(err, check.IsNil)
}

// TestVolumeCreateWrongDriver tests using wrong driver returns error.
func (suite *PouchVolumeSuite) TestVolumeCreateWrongDriver(c *check.C) {
	funcname := "TestVolumeCreateWrongDriver"

	expct := icmd.Expected{
		ExitCode: 1,
		Err:      "not found",
	}

	err := command.PouchRun("volume", "create", "--name",
		funcname, "--driver", "wrongdriver").Compare(expct)
	defer command.PouchRun("volume", "remove", funcname)

	c.Assert(err, check.IsNil)
}

// TestVolumeCreateWithLabel tests creating volume with label.
func (suite *PouchVolumeSuite) TestVolumeCreateWithLabel(c *check.C) {
	funcname := "TestVolumeCreateWithLabel"

	command.PouchRun("volume", "create", "--name", funcname, "--label", "test=foo").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "remove", funcname)
}

// TestVolumeCreateWithSelector tests creating volume with --selector.
func (suite *PouchVolumeSuite) TestVolumeCreateWithSelector(c *check.C) {
	funcname := "TestVolumeCreateWithSelector"

	command.PouchRun("volume", "create", "--name", funcname, "--selector", "test=foo").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "remove", funcname)
}

// TestVolumeCreateWithSize tests creating volume with -o opt.size=xxx.
func (suite *PouchVolumeSuite) TestVolumeCreateWithSize(c *check.C) {
	funcname := "TestVolumeCreateWithSize"

	command.PouchRun("volume", "create", "--name", funcname, "-o", "opt.size=1048576").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "remove", funcname)
}

// TestVolumeInspectFormat tests the inspect format of volume works.
func (suite *PouchVolumeSuite) TestVolumeInspectFormat(c *check.C) {
	funcname := "TestVolumeInspectFormat"

	command.PouchRun("volume", "create", "--name", funcname).Assert(c, icmd.Success)
	defer command.PouchRun("volume", "remove", funcname)

	output := command.PouchRun("volume", "inspect", funcname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	output = command.PouchRun("volume", "inspect", "-f", "{{.Name}}", funcname).Stdout()
	c.Assert(output, check.Equals, fmt.Sprintf("%s\n", funcname))
}

// TestVolumeUsingByContainer tests the inspect format of volume works.
func (suite *PouchVolumeSuite) TestVolumeUsingByContainer(c *check.C) {
	funcname := "TestVolumeUsingByContainer"

	volumeName := "volume_" + funcname
	command.PouchRun("volume", "create", "--name", volumeName).Assert(c, icmd.Success)
	command.PouchRun("run", "-d", "-v", volumeName+":/mnt", "--name", funcname, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, funcname)

	ret := command.PouchRun("volume", "rm", volumeName)
	c.Assert(ret.Stderr(), check.NotNil)

	command.PouchRun("rm", "-f", funcname).Assert(c, icmd.Success)
	command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
}

// TestVolumePluginUsingByContainer tests creating container using the plugin volume.
func (suite *PouchVolumeSuite) TestVolumePluginUsingByContainer(c *check.C) {
	funcname := "TestVolumePluginUsingByContainer"
	volumeName := "volume_" + funcname
	command.PouchRun("volume", "create", "--name", volumeName, "-d", "local-persist", "-o", "mountpoint=/data/volume1").Assert(c, icmd.Success)
	command.PouchRun("run", "-d", "-v", volumeName+":/mnt", "--name", funcname, busyboxImage, "top").Assert(c, icmd.Success)

	// delete the container.
	command.PouchRun("rm", "-f", funcname).Assert(c, icmd.Success)
	// delete the volume.
	command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
}

// TestVolumeBindReplaceMode tests the volume "direct replace(dr)" mode.
func (suite *PouchVolumeSuite) TestVolumeBindReplaceMode(c *check.C) {
	funcname := "TestVolumeBindReplaceMode"

	volumeName := "volume_" + funcname
	command.PouchRun("volume", "create", "--name", volumeName).Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("volume", "rm", volumeName)
	}()
	command.PouchRun("run", "-d", "-v", volumeName+":/mnt", "-v", volumeName+":/home:dr", "--name", funcname, busyboxImage, "top").Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("rm", "-f", funcname)
	}()

	resp, err := request.Get("/containers/" + funcname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	found := false
	for _, m := range got.Mounts {
		if m.Replace == "dr" && m.Mode == "dr" && m.Source == DefaultVolumeMountPath+"/volume_TestVolumeBindReplaceMode/home" {
			found = true
		}
	}
	c.Assert(found, check.Equals, true)
}

// TestVolumeList tests the volume list.
func (suite *PouchVolumeSuite) TestVolumeList(c *check.C) {
	var volumeName string
	expectedVolumeNames := make([]string, 0)

	for i := 1; i <= 3; i++ {
		volumeName = fmt.Sprintf("volume_TestVolumeList_%d", i)
		expectedVolumeNames = append(expectedVolumeNames, volumeName)
		command.PouchRun("volume", "create", "--name", volumeName, "-o", fmt.Sprintf("opt.size=%dg", i)).Assert(c, icmd.Success)
		defer func(volumeName string) {
			command.PouchRun("volume", "rm", volumeName)
		}(volumeName)
	}

	ret := command.PouchRun("volume", "list")
	ret.Assert(c, icmd.Success)

	lines := volumesToKV(ret.Stdout())
	for _, line := range lines {
		c.Assert(line[0], check.Equals, "local")
	}

	names := volumeNamesToSlice(ret.Stdout())
	for i, name := range names {
		c.Assert(name, check.Equals, expectedVolumeNames[i])
	}
}

// TestVolumeListOptions tests the volume list with options: size, mountpoint, quiet.
func (suite *PouchVolumeSuite) TestVolumeListOptions(c *check.C) {
	var volumeName string
	expectedVolumeNames := make([]string, 0)

	for i := 1; i <= 3; i++ {
		volumeName = fmt.Sprintf("volume_TestVolumeListOptions_%d", i)
		expectedVolumeNames = append(expectedVolumeNames, volumeName)
		command.PouchRun("volume", "create", "--name", volumeName, "-o", fmt.Sprintf("opt.size=%dg", i)).Assert(c, icmd.Success)
		defer func(volumeName string) {
			command.PouchRun("volume", "rm", volumeName)
		}(volumeName)
	}

	// test --size and --mountpoint options
	ret := command.PouchRun("volume", "list", "--size", "--mountpoint")
	ret.Assert(c, icmd.Success)

	lines := volumesToKV(ret.Stdout())
	for _, line := range lines {
		c.Assert(line[0], check.Equals, "local")
		if !strings.Contains(line[3], DefaultVolumeMountPath) {
			c.Errorf("error mount path, volume name : %s", line[1])
		}
	}

	// test --quiet options
	ret = command.PouchRun("volume", "list", "--quiet")
	ret.Assert(c, icmd.Success)

	lines = volumesToKV(ret.Stdout())
	for _, line := range lines {
		c.Assert(len(line), check.Equals, 1)
	}

	names := volumeNamesToSlice(ret.Stdout())
	for i, name := range names {
		c.Assert(name, check.Equals, expectedVolumeNames[i])
	}

	// test filter options
	volumeName = "volume_TestVolumeListOptions_4"
	command.PouchRun("volume", "create", "--name", volumeName, "--driver", "tmpfs", "--label", "test=foo").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", volumeName)

	// test name, label, driver filter separately
	filterArgs := []string{
		"name=" + volumeName,
		"label=test",
		"driver=tmpfs",
	}

	for _, args := range filterArgs {
		res := command.PouchRun("volume", "list", "--filter", args)
		res.Assert(c, icmd.Success)

		lines := volumesToKV(res.Stdout())
		c.Assert(len(lines), check.Equals, 1)
		if _, exist := lines[volumeName]; !exist {
			c.Errorf("volume filter options doesn't work, filter : ", args)
		}
	}

	// test multi volume filter
	res := command.PouchRun("volume", "list", "--filter", filterArgs[0], "--filter", filterArgs[1], "--filter", filterArgs[2])
	res.Assert(c, icmd.Success)

	lines = volumesToKV(res.Stdout())
	c.Assert(len(lines), check.Equals, 1)
	if _, exist := lines[volumeName]; !exist {
		c.Error("volume filter options doesn't work, with all filters")
	}
}

// volumesToKV parse the output of "pouch volume list" into key-value pair
func volumesToKV(volumes string) map[string][]string {
	// skip header
	lines := strings.Split(volumes, "\n")[1:]

	res := make(map[string][]string)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		items := strings.Fields(line)
		if len(items) > 1 {
			res[items[1]] = items
		} else {
			// --quiet case
			res[items[0]] = items
		}
	}
	return res
}

// volumeNamesToSlice parses volumes' name to slice.
func volumeNamesToSlice(volumes string) []string {
	lines := strings.Split(volumes, "\n")[1:]

	res := make([]string, 0)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		items := strings.Fields(line)
		if len(items) > 1 {
			res = append(res, items[1])
		} else {
			// --quiet case
			res = append(res, items[0])
		}
	}
	return res
}
