package main

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

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

// checkPidofProcess checks the process of pid and name are expected.
func checkPidofProcess(c *check.C, cname, processName, pid string) bool {
	res := command.PouchRun(
		"exec", cname,
		"ps", "axco", "pid,command")
	res.Assert(c, icmd.Success)

	output := res.Combined()
	lines := strings.Split(output, "\n")
	for _, v := range lines {
		clearStr := strings.TrimSpace(v)
		if clearStr == "" {
			continue
		}

		pidInfo := strings.SplitN(clearStr, " ", 2)
		if len(pidInfo) != 2 || pidInfo[0] != pid {
			continue
		}

		if pidInfo[1] == processName {
			return true
		}
	}

	return false
}

// checkPPid checks the ppid of process is expected.
func checkPPid(c *check.C, cname, processName, ppid string) bool {
	res := command.PouchRun(
		"exec", cname,
		"ps", "axco", "command,ppid")
	res.Assert(c, icmd.Success)

	output := res.Combined()
	lines := strings.Split(output, "\n")
	for _, v := range lines {
		clearStr := strings.TrimSpace(v)
		if clearStr == "" {
			continue
		}

		pidInfo := strings.SplitN(clearStr, " ", 2)
		if len(pidInfo) != 2 || strings.TrimSpace(pidInfo[0]) != processName {
			continue
		}

		if strings.TrimSpace(pidInfo[1]) == ppid {
			return true
		}
	}

	return false
}

// checkInitScriptWorks
func checkInitScriptWorks(c *check.C, cname string, image string, richmode string) {
	// TODO: Use bash script to get the stdin may have error, need to find a better way.

}

// TestRichContainerDumbInitWorks check the dumb-init works.
func (suite *PouchRichContainerSuite) TestRichContainerDumbInitWorks(c *check.C) {
	SkipIfFalse(c, environment.IsDumbInitExist)
	cname := "TestRichContainerDumbInitWorks"

	res := command.PouchRun("run", "-d", "--privileged", "--rich", "--rich-mode", "dumb-init", "--name", cname,
		centosImage, "sleep", "10000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	rich, err := inspectFilter(cname, ".Config.Rich")
	c.Assert(err, check.IsNil)
	c.Assert(rich, check.Equals, "true")

	richMode, err := inspectFilter(cname, ".Config.RichMode")
	c.Assert(err, check.IsNil)
	c.Assert(richMode, check.Equals, "dumb-init")

	c.Assert(checkPidofProcess(c, cname, "dumb-init", "1"), check.Equals, true)

	// stop and start could work well.
	command.PouchRun("stop", "-t", "1", cname).Assert(c, icmd.Success)
	command.PouchRun("start", cname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcess(c, cname, "dumb-init", "1"), check.Equals, true)

	// pause and unpause
	command.PouchRun("pause", cname).Assert(c, icmd.Success)
	command.PouchRun("unpause", cname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcess(c, cname, "dumb-init", "1"), check.Equals, true)
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
	cname := "TestRichContainerInitdWorks"

	ok, _ := isFileExistsInImage(centosImage, "/sbin/init", "checkinit")
	if !ok {
		c.Skip("/sbin/init doesn't exist in test image")
	}

	// --privileged is MUST required
	command.PouchRun("run", "-d", "--privileged", "--rich", "--rich-mode", "sbin-init",
		"--name", cname, centosImage, "/usr/bin/sleep 10000").Assert(c, icmd.Success)

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].Config.Rich, check.Equals, true)
	c.Assert(result[0].Config.RichMode, check.Equals, "sbin-init")

	c.Assert(checkPidofProcess(cname, "/sbin/init"), check.Equals, true)
	c.Assert(checkPPid(cname, "sleep", "1"), check.Equals, true)

	// stop and start could work well.
	command.PouchRun("stop", cname).Assert(c, icmd.Success)
	command.PouchRun("start", cname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcess(cname, "/sbin/init"), check.Equals, true)
	c.Assert(checkPPid(cname, "sleep", "1"), check.Equals, true)

	// pause and unpause
	command.PouchRun("pause", cname).Assert(c, icmd.Success)
	command.PouchRun("unpause", cname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcess(cname, "/sbin/init"), check.Equals, true)
	c.Assert(checkPPid(cname, "sleep", "1"), check.Equals, true)

	command.PouchRun("rm", "-f", cname)
}
*/

// waitSystemdPullProcess to wait util the specified process running.
// systemd will first pull the systemd processes, then pull the container
// cmd process. So we need to wait the cmd process running before check it.
func waitSystemdPullProcess(c *check.C, cname, processName string) {
	// wait systemd pull a specified process
	check := make(chan struct{})

	// check whether container started
	go func() {
		for {
			res := command.PouchRun(
				"exec", cname,
				"ps", "-ef")
			res.Assert(c, icmd.Success)

			if output := res.Combined(); strings.Contains(output, processName) {
				check <- struct{}{}
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-check:
	case <-time.After(10 * time.Second):
		c.Fatalf("failed to wait systemd pull process %s", processName)
	}

}

// TestRichContainerSystemdWorks check the systemd works.
func (suite *PouchRichContainerSuite) TestRichContainerSystemdWorks(c *check.C) {
	cname := "TestRichContainerSystemdWorks"

	ok, _ := isFileExistsInImage(centosImage, "/usr/lib/systemd/systemd", "checksysd")
	if !ok {
		c.Skip("/usr/lib/systemd/systemd doesn't exist in test image")
	}

	res := command.PouchRun("run", "-d", "--privileged", "--rich", "--rich-mode", "systemd",
		"--name", cname, centosImage, "sleep", "1000")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	rich, err := inspectFilter(cname, ".Config.Rich")
	c.Assert(err, check.IsNil)
	c.Assert(rich, check.Equals, "true")

	richMode, err := inspectFilter(cname, ".Config.RichMode")
	c.Assert(err, check.IsNil)
	c.Assert(richMode, check.Equals, "systemd")

	waitSystemdPullProcess(c, cname, "sleep")
	c.Assert(checkPidofProcess(c, cname, "systemd", "1"), check.Equals, true)
	c.Assert(checkPPid(c, cname, "sleep", "1"), check.Equals, true)

	// stop and start could work well.
	command.PouchRun("stop", "-t", "1", cname).Assert(c, icmd.Success)
	command.PouchRun("start", cname).Assert(c, icmd.Success)

	waitSystemdPullProcess(c, cname, "sleep")
	c.Assert(checkPidofProcess(c, cname, "systemd", "1"), check.Equals, true)
	c.Assert(checkPPid(c, cname, "sleep", "1"), check.Equals, true)

	// pause and unpause
	command.PouchRun("pause", cname).Assert(c, icmd.Success)
	command.PouchRun("unpause", cname).Assert(c, icmd.Success)
	c.Assert(checkPidofProcess(c, cname, "systemd", "1"), check.Equals, true)
	c.Assert(checkPPid(c, cname, "sleep", "1"), check.Equals, true)
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

// TestRunRichModeContainerWithoutRich tests running container with --rich-mode without --rich.
func (suite *PouchRichContainerSuite) TestRunRichModeContainerWithoutRich(c *check.C) {
	name := "TestRunRichModeContainerWithoutRich"

	res := command.PouchRun("run", "-d",
		"--rich-mode", "systemd",
		"--privileged",
		"--name", name, busyboxImage, "top")

	defer DelContainerForceMultyTime(c, name)
	c.Assert(res.Stderr(), check.NotNil, check.Commentf("must first enable rich mode, then specify a rich mode type"))
}

// TestRunRichModeContainerWithoutPrivileged tests running container with --rich but not --privileged.
func (suite *PouchRichContainerSuite) TestRunRichModeContainerWithoutPrivileged(c *check.C) {
	name := "TestRunRichModeContainerWithoutPrivileged"

	res := command.PouchRun("run", "-d",
		"--rich",
		"--name", name,
		busyboxImage, "top")

	defer DelContainerForceMultyTime(c, name)
	c.Assert(res.Stderr(), check.NotNil, check.Commentf("must using privileged mode when create rich container"))
}

// TestRunRichModeContainerWithWrongRichMode tests running container with --rich-mode notsupported.
func (suite *PouchRichContainerSuite) TestRunRichModeContainerWithWrongRichMode(c *check.C) {
	name := "TestRunRichModeContainerWithWrongRichMode"

	res := command.PouchRun("run", "-d",
		"--rich",
		"--rich-mode", "notsupported",
		"--privileged",
		"--name", name, busyboxImage, "top")

	defer DelContainerForceMultyTime(c, name)
	c.Assert(res.Stderr(), check.NotNil, check.Commentf("not supported rich mode"))
}
