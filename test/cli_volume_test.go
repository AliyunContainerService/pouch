package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

var (
	// DefaultVolumeMountPath defines the default volume mount path.
	DefaultVolumeMountPath = DefaultRootDir + "/volume"
)

// PouchVolumeSuite is the test suite for volume CLI.
type PouchVolumeSuite struct{}

func init() {
	check.Suite(&PouchVolumeSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchVolumeSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestVolumeWorks tests "pouch volume" work.
func (suite *PouchVolumeSuite) TestVolumeWorks(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

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

// TestVolumeWorks tests "pouch volume" work.
func (suite *PouchVolumeSuite) TestVolumeCreateLocalAndMountPoint(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	command.PouchRun("volume", "create", "--name", funcname, "--driver", "local", "-o", "mount=/tmp").Assert(c, icmd.Success)
	output := command.PouchRun("volume", "inspect", funcname).Stdout()
	if !strings.Contains(output, "local") {
		c.Errorf("failed to get the backend driver, expect:local, acturally: %s", output)
	}

	if !strings.Contains(output, "/tmp/"+funcname) {
		c.Errorf("failed to get the mountpoint, expect:/tmp/%s, acturally: %s", funcname, output)
	}

	command.PouchRun("volume", "remove", funcname).Assert(c, icmd.Success)
}

// TestVolumeCreateLocalDriverAndSpecifyMountPoint tests "pouch volume create" works.
func (suite *PouchVolumeSuite) TestVolumeCreateLocalDriverAndSpecifyMountPoint(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	command.PouchRun("volume", "create", "--name", funcname, "--driver", "local", "-o", "mount=/tmp").Assert(c, icmd.Success)
	output := command.PouchRun("volume", "inspect", funcname).Stdout()
	if !strings.Contains(output, "local") {
		c.Errorf("failed to get the backend driver, expect:local, acturally: %s", output)
	}

	if !strings.Contains(output, "/tmp/"+funcname) {
		c.Errorf("failed to get the mountpoint, expect:/tmp/%s, acturally: %s", funcname, output)
	}

	command.PouchRun("volume", "remove", funcname).Assert(c, icmd.Success)
}

// TestVolumeCreateWithMountPointExitsFile tests when MountPoint is an existing file, returns error.
func (suite *PouchVolumeSuite) TestVolumeCreateWithMountPointExitsFile(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	expct := icmd.Expected{
		ExitCode: 1,
		Err:      "mount path is not a dir",
	}

	icmd.RunCommand("touch", "/tmp/"+funcname)
	err := command.PouchRun("volume", "create", "--name", funcname, "--driver", "local", "-o", "mount=/tmp").Compare(expct)
	c.Assert(err, check.IsNil)

	command.PouchRun("volume", "remove", funcname)
}

// TestVolumeCreateWrongDriver tests using wrong driver returns error.
func (suite *PouchVolumeSuite) TestVolumeCreateWrongDriver(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	expct := icmd.Expected{
		ExitCode: 1,
		Err:      "not found",
	}

	err := command.PouchRun("volume", "create", "--name", funcname, "--driver", "wrongdriver").Compare(expct)
	c.Assert(err, check.IsNil)

	command.PouchRun("volume", "remove", funcname)
}

// TestVolumeCreateWithLabel tests creating volume with label.
func (suite *PouchVolumeSuite) TestVolumeCreateWithLabel(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	command.PouchRun("volume", "create", "--name", funcname, "--label", "test=foo").Assert(c, icmd.Success)
	command.PouchRun("volume", "remove", funcname)
}

// TestVolumeCreateWithSelector tests creating volume with --selector.
func (suite *PouchVolumeSuite) TestVolumeCreateWithSelector(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	command.PouchRun("volume", "create", "--name", funcname, "--selector", "test=foo").Assert(c, icmd.Success)
	command.PouchRun("volume", "remove", funcname)
}

// TestVolumeInspectFormat tests the inspect format of volume works.
func (suite *PouchVolumeSuite) TestVolumeInspectFormat(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

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
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	volumeName := "volume_" + funcname
	command.PouchRun("volume", "create", "--name", volumeName).Assert(c, icmd.Success)
	command.PouchRun("run", "-d", "-v", volumeName+":/mnt", "--name", funcname, busyboxImage, "top").Assert(c, icmd.Success)

	ret := command.PouchRun("volume", "rm", volumeName)
	c.Assert(ret.Error, check.NotNil)

	command.PouchRun("rm", "-f", funcname).Assert(c, icmd.Success)
	command.PouchRun("volume", "rm", volumeName).Assert(c, icmd.Success)
}

// TestVolumeBindReplaceMode tests the volume "direct replace(dr)" mode.
func (suite *PouchVolumeSuite) TestVolumeBindReplaceMode(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	volumeName := "volume_" + funcname
	command.PouchRun("volume", "create", "--name", volumeName).Assert(c, icmd.Success)
	command.PouchRun("run", "-d", "-v", volumeName+":/mnt", "-v", volumeName+":/home:dr", "--name", funcname, busyboxImage, "top").Assert(c, icmd.Success)
	defer func() {
		command.PouchRun("volume", "rm", volumeName)
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
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	volumeName := "volume_" + funcname
	volumeName1 := "volume_" + funcname + "_1"
	command.PouchRun("volume", "create", "--name", volumeName1, "-o", "size=1g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", volumeName1)

	volumeName2 := "volume_" + funcname + "_2"
	command.PouchRun("volume", "create", "--name", volumeName2, "-o", "size=2g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", volumeName2)

	volumeName3 := "volume_" + funcname + "_3"
	command.PouchRun("volume", "create", "--name", volumeName3, "-o", "size=3g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", volumeName3)

	ret := command.PouchRun("volume", "list")
	ret.Assert(c, icmd.Success)

	for _, line := range strings.Split(ret.Stdout(), "\n") {
		if strings.Contains(line, volumeName) {
			if !strings.Contains(line, "local") {
				c.Errorf("list result have no driver or name or size or mountpoint, line: %s", line)
				break
			}
		}
	}
}

// TestVolumeList tests the volume list with options: size and mountpoint.
func (suite *PouchVolumeSuite) TestVolumeListOptions(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	volumeName := "volume_" + funcname
	volumeName1 := "volume_" + funcname + "_1"
	command.PouchRun("volume", "create", "--name", volumeName1, "-o", "size=1g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", volumeName1)

	volumeName2 := "volume_" + funcname + "_2"
	command.PouchRun("volume", "create", "--name", volumeName2, "-o", "size=2g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", volumeName2)

	volumeName3 := "volume_" + funcname + "_3"
	command.PouchRun("volume", "create", "--name", volumeName3, "-o", "size=3g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", volumeName3)

	ret := command.PouchRun("volume", "list", "--size", "--mountpoint")
	ret.Assert(c, icmd.Success)

	for _, line := range strings.Split(ret.Stdout(), "\n") {
		if strings.Contains(line, volumeName) {
			if !strings.Contains(line, "local") ||
				!strings.Contains(line, "g") ||
				!strings.Contains(line, DefaultVolumeMountPath) {
				c.Errorf("list result have no driver or name or size or mountpoint, line: %s", line)
				break
			}
		}
	}
}
