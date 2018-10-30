package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

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

	res := command.PouchRun("run", "-d", "--rich", "--rich-mode", "dumb-init", "--name", funcname,
		busyboxImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, funcname)
	res.Assert(c, icmd.Success)

	rich, err := inspectFilter(funcname, ".Config.Rich")
	c.Assert(err, check.IsNil)
	c.Assert(rich, check.Equals, "true")

	richMode, err := inspectFilter(funcname, ".Config.RichMode")
	c.Assert(err, check.IsNil)
	c.Assert(richMode, check.Equals, "dumb-init")

	c.Assert(checkPidofProcessIsOne(funcname, "dumb-init"), check.Equals, true)

	// stop and start could work well.
	command.PouchRun("stop", funcname).Assert(c, icmd.Success)
	command.PouchRun("start", funcname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcessIsOne(funcname, "dumb-init"), check.Equals, true)

	// pause and unpause
	command.PouchRun("pause", funcname).Assert(c, icmd.Success)
	command.PouchRun("unpause", funcname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcessIsOne(funcname, "dumb-init"), check.Equals, true)
}

// TestRichContainerWrongArgs check the wrong args of rich container.
func (suite *PouchRichContainerSuite) TestRichContainerDumbInitWrongArgs(c *check.C) {
	SkipIfFalse(c, environment.IsDumbInitExist)

	// TODO
	// Don't add '--rich' when use other rich container related options should fail.

}

// TestRichContainerSbinInitWorks check the initd works.
/*
Comment the test (Ace-Tang).
related issue : https://github.com/alibaba/pouch/issues/960
related pr: https://github.com/alibaba/pouch/pull/1128
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
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].Config.Rich, check.Equals, true)
	c.Assert(result[0].Config.RichMode, check.Equals, "sbin-init")

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
*/

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

	res := command.PouchRun("run", "-d", "--privileged", "--rich", "--rich-mode", "systemd",
		"--name", funcname, centosImage, "/usr/bin/sleep 1000")
	defer DelContainerForceMultyTime(c, funcname)
	res.Assert(c, icmd.Success)

	rich, err := inspectFilter(funcname, ".Config.Rich")
	c.Assert(err, check.IsNil)
	c.Assert(rich, check.Equals, "true")

	richMode, err := inspectFilter(funcname, ".Config.RichMode")
	c.Assert(err, check.IsNil)
	c.Assert(richMode, check.Equals, "systemd")

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
}

// TestRichContainerUpdateEnvFile check update env and env file successfully.
func (suite *PouchRichContainerSuite) TestRichContainerUpdateEnvFile(c *check.C) {
	name := "TestRichContainerUpdateEnvFile"

	res := command.PouchRun("run", "-d", "--name", name, centosImage, "sleep", "100")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	command.PouchRun("update", "--env", "foo=bar", name).Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	if !utils.StringInSlice(result[0].Config.Env, "foo=bar") {
		c.Errorf("expect 'foo=bar' in container env, but got: %v", result[0].Config.Env)
	}

	// check foo="bar" should exist in /etc/profile.d/pouchenv.sh
	output = command.PouchRun("exec", name, "cat", "/etc/profile.d/pouchenv.sh").Stdout()
	if !strings.Contains(output, "foo=\"bar\"") {
		c.Errorf("failed to update /etc/profile.d/pouchenv.sh, got content: %s", output)
	}

	// check env foo=bar update successfully.
	output = command.PouchRun("exec", name, "env").Stdout()
	if !strings.Contains(output, "foo=bar") {
		c.Errorf("failed to update env, got content: %s", output)
	}
}
