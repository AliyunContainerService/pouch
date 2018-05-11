package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIContainerLogsSuite is the test suite for container logs API.
type APIContainerLogsSuite struct {
	interestingFunctions []string
}

func init() {
	check.Suite(&APIContainerLogsSuite{
		interestingFunctions: []string{
			"server.logsContainer",
			"(*ContainerManager).ReadLogMessages",
			"jsonfile.(*JSONLogFile).read",
		},
	})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerLogsSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestNoSuchContainer tests a container that doesn't exits return error.
func (suite *APIContainerLogsSuite) TestNoSuchContainer(c *check.C) {
	resp, err := request.Get("/containers/nosuchcontainerxxx/logs")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, http.StatusNotFound)
}

// TestNoShowStdoutAndShowStderr tests logs API without ShowStderr and
// ShowStdout should return 401.
func (suite *APIContainerLogsSuite) TestNoShowStdoutAndShowStderr(c *check.C) {
	name := "logs_without_one_of_showstdout_and_showstderr"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "ls").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	resp, err := request.Get(fmt.Sprintf("/containers/%s/logs", name))
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, http.StatusBadRequest)

	DelContainerForceOk(c, name)
}

// TestStdout tests stdout stream.
func (suite *APIContainerLogsSuite) TestStdout(c *check.C) {
	name := "logs_stdout_stream"
	command.PouchRun("run", "-t", "--name", name, busyboxImage, "echo", "hello").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	allLogs := suite.syncReadLog(c, name, map[string]string{"stdout": "1"})
	c.Assert(len(allLogs), check.Equals, 1)
	c.Assert(strings.TrimSpace(allLogs[0]), check.Equals, "hello")
}

// TestTimestamp tests stdout stream with timestamp.
func (suite *APIContainerLogsSuite) TestTimestamp(c *check.C) {
	name := "logs_stdout_stream_with_timestamp"
	command.PouchRun("run", "-t", "--name", name, busyboxImage, "echo", "hello").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	allLogs := suite.syncReadLog(c, name, map[string]string{"stdout": "1", "timestamps": "1"})
	c.Assert(len(allLogs), check.Equals, 1)

	_, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[0], " ")[0])
	c.Assert(err, check.IsNil)
	c.Assert(strings.TrimSpace(strings.Split(allLogs[0], " ")[1]), check.Equals, "hello")
}

// TestTails tests limit log lines.
func (suite *APIContainerLogsSuite) TestTails(c *check.C) {
	name := "logs_stdout_stream_with_tails"
	command.PouchRun("run", "-t", "--name", name, busyboxImage, "sh", "-c", "for i in $(seq 1 3); do echo hi$i; done").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	allLogs := suite.syncReadLog(c, name, map[string]string{"stdout": "1", "tail": "2"})
	c.Assert(len(allLogs), check.Equals, 2)
	for i := range allLogs {
		c.Assert(allLogs[i], check.Equals, fmt.Sprintf("hi%d", i+2))
	}
}

// TestSinceAndUntil tests limit log lines by since and until query.
func (suite *APIContainerLogsSuite) TestSinceAndUntil(c *check.C) {
	name := "logs_stdout_stream_with_tails"
	command.PouchRun("run", "-t", "--name", name, busyboxImage, "sh", "-c", "for i in $(seq 1 3); do echo hi$i; sleep 1; done").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	allLogs := suite.syncReadLog(c, name, map[string]string{"stdout": "1", "timestamps": "1"})
	c.Assert(len(allLogs), check.Equals, 3)

	// get the since and until time
	sinceTime, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[1], " ")[0])
	c.Assert(err, check.IsNil)

	untilTime, err := time.Parse(utils.TimeLayout, strings.Split(allLogs[2], " ")[0])
	c.Assert(err, check.IsNil)
	untilTime = untilTime.Add(-100 * time.Nanosecond)

	allLogs = suite.syncReadLog(c, name, map[string]string{
		"stdout": "1",
		"since":  fmt.Sprintf("%d.%09d", sinceTime.Unix(), sinceTime.UnixNano()),
		"until":  fmt.Sprintf("%d.%09d", untilTime.Unix(), untilTime.UnixNano()),
	})

	c.Assert(len(allLogs), check.Equals, 1)
}

func (suite *APIContainerLogsSuite) TestCheckLeakByFollowing(c *check.C) {
	name := "logs_check_leak_by_following"
	command.PouchRun("run", "-d", "--name", name, busyboxImage, "sh", "-c", "while true; do sleep 1; done").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	// follow one second and check the goroutine leak
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	request.Get(fmt.Sprintf("/containers/%s/logs", name),
		request.WithContext(ctx),
		request.WithQuery(
			url.Values(map[string][]string{
				"stdout": {"1"},
				"follow": {"1"},
			}),
		),
	)

	// waiting the closeNotify to clean the goroutine
	<-time.After(2 * time.Second)
	for _, g := range suite.fetchCurrentStack(c) {
		for _, fun := range suite.interestingFunctions {
			if strings.Contains(g.Stacktrace, fun) {
				c.Fatalf("the goroutine should be clean:\n %v", g.Stacktrace)
			}
		}
	}
}

// syncReadLog will read all the log content from pouchd.
func (suite *APIContainerLogsSuite) syncReadLog(c *check.C, id string, query map[string]string) []string {
	q := url.Values{}
	for k, v := range query {
		// skip the follow mode
		if k == "follow" {
			continue
		}
		q.Set(k, v)
	}

	resp, err := request.Get(fmt.Sprintf("/containers/%s/logs", id), request.WithQuery(q))
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, http.StatusOK)

	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	return strings.Split(strings.TrimSpace(string(out)), "\n")
}

func (suite *APIContainerLogsSuite) fetchCurrentStack(c *check.C) []*util.Goroutine {
	resp, err := request.Debug("/debug/pprof/goroutine?debug=2")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, http.StatusOK)

	out, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)

	return util.GoroutinesFromStack(out)
}
