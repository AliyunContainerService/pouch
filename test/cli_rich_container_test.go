package main

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"encoding/json"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRichContainerSuite is the test suite fo rich container related CLI.
type PouchRichContainerSuite struct{}

func init() {
	check.Suite(&PouchRichContainerSuite{})
}

var centosImage string

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRichContainerSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	SkipIfFalse(c, environment.IsRuncVersionSupportRichContianer)

	PullImage(c, busyboxImage)

	// Use image from AliYun on AliOS.
	if environment.IsAliKernel() {
		centosImage = "reg.docker.alibaba-inc.com/alibase/alios7u2:latest"
	} else {
		centosImage = "registry.hub.docker.com/library/centos:latest"
	}
	command.PouchRun("pull", centosImage).Assert(c, icmd.Success)
}

// TearDownSuite does common cleanup in the end of each test suite.
func (suite *PouchRichContainerSuite) TearDownSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	SkipIfFalse(c, environment.IsRuncVersionSupportRichContianer)

	command.PouchRun("rmi", centosImage)
}

// isFileExistsInImage checks if the file exists in given image.
func isFileExistsInImage(image string, file string, cname string) (bool, error) {
	if image == "" || file == "" || cname == "" {
		return false, errors.New("input args is nil")
	}

	// check the existence of /sbin/init in image
	expect := icmd.Expected{
		ExitCode: 0,
		Out:      "Access",
	}
	err := command.PouchRun("run", "--name", cname, image, "stat", file).Compare(expect)
	command.PouchRun("rm", "-f", cname)

	return err == nil, nil
}

// checkPidofProcessIsOne checks the process of pid 1 is expected.
func checkPidofProcessIsOne(cname string, p string) bool {
	expect := icmd.Expected{
		ExitCode: 0,
		Out:      "1",
	}

	err := command.PouchRun("exec", cname,
		"ps", "-ef", "| grep", p, "| awk '{print $1}'").Compare(expect)
	if err != nil {
		fmt.Printf("err=%s\n", err)
	}
	return err == nil
}

// checkPPid checks the ppid of process is expected.
func checkPPid(cname string, p string, ppid string) bool {
	expect := icmd.Expected{
		ExitCode: 0,
		Out:      ppid,
	}

	err := command.PouchRun("exec", cname,
		"ps", "-ef", "|grep", p, "|awk '{print $3}'").Compare(expect)
	if err != nil {
		fmt.Printf("err=%s\n", err)
	}

	return err == nil
}

// checkInitScriptWorks
func checkInitScriptWorks(c *check.C, cname string, image string, richmode string) {
	// TODO: Use bash script to get the stdin may have error, need to find a better way.

}

// TestRichContainerDumbInitWorks check the dumb-init works.
func (suite *PouchRichContainerSuite) TestRichContainerDumbInitWorks(c *check.C) {
	SkipIfFalse(c, environment.IsDumbInitExist)
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	command.PouchRun("run", "-d", "--rich", "--rich-mode", "dumb-init", "--name", funcname,
		busyboxImage, "sleep", "10000").Assert(c, icmd.Success)

	output := command.PouchRun("inspect", funcname).Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.Config.Rich, check.Equals, true)
	c.Assert(result.Config.RichMode, check.Equals, "dumb-init")

	c.Assert(checkPidofProcessIsOne(funcname, "dumb-init"), check.Equals, true)

	// stop and start could work well.
	command.PouchRun("stop", funcname).Assert(c, icmd.Success)
	command.PouchRun("start", funcname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcessIsOne(funcname, "dumb-init"), check.Equals, true)

	// pause and unpause
	command.PouchRun("pause", funcname).Assert(c, icmd.Success)
	command.PouchRun("unpause", funcname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcessIsOne(funcname, "dumb-init"), check.Equals, true)

	command.PouchRun("rm", "-f", funcname)
}

// TestRichContainerWrongArgs check the wrong args of rich container.
func (suite *PouchRichContainerSuite) TestRichContainerDumbInitWrongArgs(c *check.C) {
	SkipIfFalse(c, environment.IsDumbInitExist)

	// TODO
	// Don't add '--rich' when use other rich container related options should fail.

}

// TestRichContainerSbinInitWorks check the initd works.
func (suite *PouchRichContainerSuite) TestRichContainerInitdWorks(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	ok, _ := isFileExistsInImage(centosImage, "/sbin/init", "checkinit")
	if !ok {
		c.Skip("/sbin/init doesn't exist in test image")
	}

	// --privileged is MUST required
	command.PouchRun("run", "-d", "--privileged", "--rich", "--rich-mode", "sbin-init",
		"--name", funcname, centosImage, "/usr/bin/sleep 10000").Assert(c, icmd.Success)

	output := command.PouchRun("inspect", funcname).Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.Config.Rich, check.Equals, true)
	c.Assert(result.Config.RichMode, check.Equals, "sbin-init")

	c.Assert(checkPidofProcessIsOne(funcname, "/sbin/init"), check.Equals, true)
	c.Assert(checkPPid(funcname, "sleep", "1"), check.Equals, true)

	// stop and start could work well.
	command.PouchRun("stop", funcname).Assert(c, icmd.Success)
	command.PouchRun("start", funcname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcessIsOne(funcname, "/sbin/init"), check.Equals, true)
	c.Assert(checkPPid(funcname, "sleep", "1"), check.Equals, true)

	// pause and unpause
	command.PouchRun("pause", funcname).Assert(c, icmd.Success)
	command.PouchRun("unpause", funcname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcessIsOne(funcname, "/sbin/init"), check.Equals, true)
	c.Assert(checkPPid(funcname, "sleep", "1"), check.Equals, true)

	command.PouchRun("rm", "-f", funcname)
}

// TestRichContainerSystemdWorks check the systemd works.
func (suite *PouchRichContainerSuite) TestRichContainerSystemdWorks(c *check.C) {
	// TODO: uncomment it
	c.Skip("skip this flaky test")

	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	ok, _ := isFileExistsInImage(centosImage, "/usr/lib/systemd/systemd", "checksysd")
	if !ok {
		c.Skip("/usr/lib/systemd/systemd doesn't exist in test image")
	}

	command.PouchRun("run", "-d", "--privileged", "--rich", "--rich-mode", "systemd",
		"--name", funcname, centosImage, "/usr/bin/sleep 1000").Assert(c, icmd.Success)

	output := command.PouchRun("inspect", funcname).Stdout()
	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.Config.Rich, check.Equals, true)
	c.Assert(result.Config.RichMode, check.Equals, "systemd")

	c.Assert(checkPidofProcessIsOne(funcname, "/usr/lib/systemd/systemd"), check.Equals, true)
	c.Assert(checkPPid(funcname, "sleep", "1"), check.Equals, true)

	// stop and start could work well.
	command.PouchRun("stop", funcname).Assert(c, icmd.Success)
	command.PouchRun("start", funcname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcessIsOne(funcname, "/usr/lib/systemd/systemd"), check.Equals, true)
	c.Assert(checkPPid(funcname, "sleep", "1"), check.Equals, true)

	// pause and unpause
	command.PouchRun("pause", funcname).Assert(c, icmd.Success)
	command.PouchRun("unpause", funcname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcessIsOne(funcname, "/usr/lib/systemd/systemdd"), check.Equals, true)
	c.Assert(checkPPid(funcname, "sleep", "1"), check.Equals, true)

	command.PouchRun("rm", "-f", funcname)
}
