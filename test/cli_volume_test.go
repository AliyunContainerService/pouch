package main

import (
	"runtime"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchVolumeSuite is the test suite fo help CLI.
type PouchVolumeSuite struct{}

func init() {
	check.Suite(&PouchVolumeSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchVolumeSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
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
	icmd.RunCommand("stat", "/mnt/local/"+funcname).Compare(expct)
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
	command.PouchRun("volume", "create", "--name", funcname, "--driver", "local", "-o", "mount=/tmp").Compare(expct)
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

	command.PouchRun("volume", "create", "--name", funcname, "--driver", "wrongdriver").Compare(expct)
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
