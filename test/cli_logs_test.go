package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchLogsSuite is the test suite for logs CLI.
type PouchLogsSuite struct{}

func init() {
	check.Suite(&PouchLogsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchLogsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TestCreatedContainerLogIsEmpty tests logs for created container.
func (suite *PouchLogsSuite) TestCreatedContainerLogIsEmpty(c *check.C) {
	cname := "TestCLILogs_EmptyLogInCreatedContainer"

	command.PouchRun("create", "--name", cname, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	res := command.PouchRun("logs", cname)
	res.Assert(c, icmd.Success)
	c.Assert(res.Combined(), check.Equals, "")
}

func (suite *PouchLogsSuite) TestLogsSeparateStderr(c *check.C) {
	cname := "TestLogsSeparateStderr"
	msg := "stderr_log"
	command.PouchRun("run", "-d", "--name", cname, busyboxImage, "sh", "-c", fmt.Sprintf("echo %s 1>&2", msg)).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	command.PouchRun("logs", cname).Assert(c, icmd.Expected{
		Out: "",
		Err: msg,
	})
}

func (suite *PouchLogsSuite) TestLogsStderrInStdout(c *check.C) {
	cname := "TestLogsStderrInStdout"
	msg := "stderr_log"
	command.PouchRun("run", "-d", "-t", "--name", cname, busyboxImage, "sh", "-c", fmt.Sprintf("echo %s 1>&2", msg)).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	command.PouchRun("logs", cname).Assert(c, icmd.Expected{
		Out: msg,
		Err: "",
	})
}

// TestSinceAndUntil tests the since and until.
func (suite *PouchLogsSuite) TestSinceAndUntil(c *check.C) {
	cname := "TestCLILogs_Since_and_Until"
	totalLine := 5

	command.PouchRun(
		"run",
		"-t",
		"--name", cname, busyboxImage,
		"sh", "-c", fmt.Sprintf("for i in $(seq 1 %v); do echo hello$i; sleep 1; done;", totalLine),
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	allLogs := suite.syncLogs(c, cname, "--timestamps")
	c.Assert(len(allLogs), check.Equals, totalLine)

	// get the since and until time
	sinceTime, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[2], " ")[0])
	c.Assert(err, check.IsNil)

	untilTime, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[3], " ")[0])
	c.Assert(err, check.IsNil)
	untilTime = untilTime.Add(-200 * time.Nanosecond)

	allLogs = suite.syncLogs(c, cname,
		"--since", fmt.Sprintf("%d", sinceTime.UnixNano()),
		"--until", fmt.Sprintf("%d", untilTime.UnixNano()))
	c.Assert(len(allLogs), check.Equals, 1)
}

// TestTimestamp tests the timestamps flag.
func (suite *PouchLogsSuite) TestTimestamp(c *check.C) {
	cname := "TestCLILogs_timestamp"

	command.PouchRun(
		"run",
		"-t",
		"--name", cname,
		busyboxImage,
		"echo", "hello",
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	allLogs := suite.syncLogs(c, cname, "--timestamps")
	c.Assert(len(allLogs), check.Equals, 1)

	_, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[0], " ")[0])
	c.Assert(err, check.IsNil)
	c.Assert(strings.TrimSpace(strings.Split(allLogs[0], " ")[1]), check.Equals, "hello")
}

// TestTailMode tests follow mode.
func (suite *PouchLogsSuite) TestTailLine(c *check.C) {
	cname := "TestCLILogs_tail_line"

	totalLine := 100

	command.PouchRun(
		"run",
		"-t",
		"--name", cname,
		busyboxImage,
		"sh", "-c", fmt.Sprintf("for i in $(seq 1 %v); do echo hello-$i; done;", totalLine),
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	for _, tc := range []struct {
		input    string
		expected int
	}{
		{"1000", totalLine},
		{"100", totalLine},
		{"67", 67},
		{"5", 5},
		{"0", totalLine},
		{"-1", totalLine},
		{"wronglinenumber", totalLine},
	} {
		allLogs := suite.syncLogs(c, cname, "--tail", tc.input)
		c.Assert(allLogs, check.HasLen, tc.expected)
	}
}

// TestFollowMode tests follow mode.
func (suite *PouchLogsSuite) TestFollowMode(c *check.C) {
	cname := "TestCLILogs_follow_mode"

	command.PouchRun(
		"run",
		"-d",
		"--name", cname,
		busyboxImage,
		"sh", "-c", "for i in $(seq 1 3); do sleep 2; echo hello; done;",
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	waitCh := make(chan error)
	go func() {
		waitCh <- command.PouchRun("logs", "-f", cname).Error
		close(waitCh)
	}()

	select {
	case err := <-waitCh:
		c.Assert(err, check.IsNil)
	case <-time.After(15 * time.Second):
		c.Fatal("pouch logs -f may hanged")
	}
}

// TestLogsOpt tests if log options could work.
func (suite *PouchLogsSuite) TestLogsOpt(c *check.C) {
	cname := "TestCLILogs_LogsOpt"
	command.PouchRun(
		"run",
		"--log-opt", "env=test",
		"--name", cname,
		busyboxImage,
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	cnameOfUnsupported := "TestCLILogs_LogsOpt_Unsupported"
	result := command.PouchRun(
		"run",
		"--log-opt", "env1=test",
		"--name", cnameOfUnsupported,
		busyboxImage,
	)
	c.Assert(result.Error, check.NotNil)
	defer DelContainerForceMultyTime(c, cnameOfUnsupported)
}

// TestLogsWithDetails tests details opt.
func (suite *PouchLogsSuite) TestLogsWithDetails(c *check.C) {
	cname := "TestLogsWithDetails"

	res := command.PouchRun("run", "--name", cname,
		"--label", "foo=bar",
		"-e", "baz=qux",
		"--log-opt", "labels=foo",
		"--log-opt", "env=baz",
		busyboxImage, "echo", "hello")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	out := suite.syncLogs(c, cname, "--details")
	c.Assert(len(out), check.Equals, 1)
	logFields := strings.Split(out[0], " ")

	details := strings.Split(logFields[0], ",")

	c.Assert(len(details), check.Equals, 2)
	c.Assert(details[0], check.Equals, "baz=qux")
	c.Assert(details[1], check.Equals, "foo=bar")

	cnameOfEmptyDetails := "TestLogsWithEmptyDetails"

	command.PouchRun("run",
		"--name", cnameOfEmptyDetails,
		busyboxImage,
		"echo", "hello",
	).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cnameOfEmptyDetails)

	logs := suite.syncLogs(c, cnameOfEmptyDetails, "--details")
	c.Assert(len(logs), check.Equals, 1)
	c.Assert(logs[0], check.Equals, "hello")
}

func (suite *PouchLogsSuite) syncLogs(c *check.C, cname string, flags ...string) []string {
	args := append([]string{"logs"}, flags...)

	res := command.PouchRun(append(args, cname)...)
	res.Assert(c, icmd.Success)

	return strings.Split(strings.TrimSpace(string(res.Combined())), "\n")
}

// TestSetLogPathInDaemon tests set root-dir of container log path in daemon config
func (suite *PouchLogsSuite) TestSetLogPathInDaemon(c *check.C) {
	logRootDir := "/tmp/TestSetLogsPathInDaemon"
	defer os.RemoveAll(logRootDir)

	dcfg, err := StartDefaultDaemon("--log-opt", fmt.Sprintf("root-dir=%s", logRootDir))
	c.Assert(err, check.IsNil)

	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
	c.Assert(result.ExitCode, check.Equals, 0)
	// clean busybox image
	defer RunWithSpecifiedDaemon(dcfg, "rmi", busyboxImage)

	cname := "TestSetLogsPathInDaemonCnt"
	result = RunWithSpecifiedDaemon(dcfg, "create", "--net", "none", "--name", cname, busyboxImage, "sh", "-c", "echo hello")
	c.Assert(result.ExitCode, check.Equals, 0)

	defer RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname)

	result = RunWithSpecifiedDaemon(dcfg, "inspect", "-f", "{{.ID}}", cname)
	c.Assert(result.ExitCode, check.Equals, 0)

	cid := strings.TrimSpace(result.Stdout())

	result = RunWithSpecifiedDaemon(dcfg, "start", cname)
	c.Assert(result.ExitCode, check.Equals, 0)

	out, err := ioutil.ReadFile(filepath.Join(logRootDir, cid, "json.log"))
	c.Assert(err, check.IsNil)
	c.Assert(strings.Contains(string(out), "hello"), check.Equals, true)

	result = RunWithSpecifiedDaemon(dcfg, "inspect", "-f", "{{.LogPath}}", cname)
	c.Assert(result.ExitCode, check.Equals, 0)
	logPath := strings.TrimSpace(result.Stdout())
	c.Assert(logPath, check.Equals, filepath.Join(logRootDir, cid, "json.log"))

	result = RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname)
	c.Assert(result.ExitCode, check.Equals, 0)

	_, err = os.Stat(filepath.Join(logRootDir, cid))
	c.Assert(os.IsNotExist(err), check.Equals, true)

	// case: set root-dir in container config
	containerCfgLogRootDir := "/tmp/TestSetLogsPathInDaemon_containerCfg"
	defer os.RemoveAll(containerCfgLogRootDir)

	cname2 := "TestSetLogsPathInDaemonCnt_v2"

	result = RunWithSpecifiedDaemon(dcfg, "create", "--net", "none", "--log-opt", fmt.Sprintf("root-dir=%s", containerCfgLogRootDir), "--name", cname2, busyboxImage, "sh", "-c", "echo world")
	c.Assert(result.ExitCode, check.Equals, 0)

	defer RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname2)

	result = RunWithSpecifiedDaemon(dcfg, "inspect", "-f", "{{.ID}}", cname2)
	c.Assert(result.ExitCode, check.Equals, 0)

	cid = strings.TrimSpace(result.Stdout())

	result = RunWithSpecifiedDaemon(dcfg, "start", cname2)
	c.Assert(result.ExitCode, check.Equals, 0)

	out, err = ioutil.ReadFile(filepath.Join(containerCfgLogRootDir, cid, "json.log"))
	c.Assert(err, check.IsNil)
	c.Assert(strings.Contains(string(out), "world"), check.Equals, true)

	result = RunWithSpecifiedDaemon(dcfg, "inspect", "-f", "{{.LogPath}}", cname2)
	c.Assert(result.ExitCode, check.Equals, 0)
	logPath = strings.TrimSpace(result.Stdout())
	c.Assert(logPath, check.Equals, filepath.Join(containerCfgLogRootDir, cid, "json.log"))

	result = RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname2)
	c.Assert(result.ExitCode, check.Equals, 0)

	_, err = os.Stat(filepath.Join(containerCfgLogRootDir, cid))
	c.Assert(os.IsNotExist(err), check.Equals, true)
}

// TestSetLogPathByContainerConfig tests set root-dir of container log path in container config
func (suite *PouchLogsSuite) TestSetLogPathByContainerConfig(c *check.C) {
	logRootDir := "/tmp/TestSetLogPathByContainerConfig"
	defer os.RemoveAll(logRootDir)

	cname := "TestSetLogPathByContainerConfigCnt"
	result := command.PouchRun("create", "--net", "none", "--log-opt", fmt.Sprintf("root-dir=%s", logRootDir), "--name", cname, busyboxImage, "sh", "-c", "echo hello")
	c.Assert(result.ExitCode, check.Equals, 0)

	defer command.PouchRun("rm", "-f", cname)

	result = command.PouchRun("inspect", "-f", "{{.ID}}", cname)
	c.Assert(result.ExitCode, check.Equals, 0)

	cid := strings.TrimSpace(result.Stdout())

	result = command.PouchRun("start", cname)
	c.Assert(result.ExitCode, check.Equals, 0)

	out, err := ioutil.ReadFile(filepath.Join(logRootDir, cid, "json.log"))
	c.Assert(err, check.IsNil)
	c.Assert(strings.Contains(string(out), "hello"), check.Equals, true)

	result = command.PouchRun("inspect", "-f", "{{.LogPath}}", cname)
	c.Assert(result.ExitCode, check.Equals, 0)
	logPath := strings.TrimSpace(result.Stdout())
	c.Assert(logPath, check.Equals, filepath.Join(logRootDir, cid, "json.log"))

	result = command.PouchRun("rm", "-f", cname)
	c.Assert(result.ExitCode, check.Equals, 0)

	_, err = os.Stat(filepath.Join(logRootDir, cid))
	c.Assert(os.IsNotExist(err), check.Equals, true)
}
