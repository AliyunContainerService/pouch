package main

import (
	"fmt"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerMetricsSuite is the test suite for container metrics API.
type APIContainerMetricsSuite struct {
	cname string
}

func init() {
	check.Suite(&APIContainerMetricsSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerMetricsSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

}

// SetUpSuite does common setup in the beginning of each suite .
func (suite *APIContainerMetricsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	suite.cname = "TestMetricsContainer"
	PullImage(c, busyboxImage)
}

// TearDownSuite run after each suite to do cleanup work for the whole suite.
func (suite *APIContainerMetricsSuite) TearDownSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	request.Delete("/containers/" + suite.cname + "?force=1")
}

// TestContainerMetrics test metrics of container.
func (suite *APIContainerMetricsSuite) TestContainerMetrics(c *check.C) {
	cname := suite.cname

	suite.checkAction(c, cname, "create")
	suite.checkAction(c, cname, "start")
	suite.checkAction(c, cname, "stop")
	suite.checkAction(c, cname, "delete")
}

func (suite *APIContainerMetricsSuite) checkAction(c *check.C, cname string, label string) {
	key := fmt.Sprintf(`engine_daemon_container_actions_total{action="%s"}`, label)
	keySuccess := fmt.Sprintf(`engine_daemon_container_success_actions_total{action="%s"}`, label)
	countBefore, countSuccessBefore := GetMetric(c, key, keySuccess)

	switch label {
	case "create":
		CreateBusyboxContainerOk(c, cname)
	case "start":
		StartContainerOk(c, cname)
	case "stop":
		StopContainerOk(c, cname)
	case "delete":
		resp, err := request.Delete("/containers/" + cname)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 204)
	default:
		fmt.Println("error")
		c.Fatal("error")
	}

	count, successCount := GetMetric(c,
		key,
		keySuccess)
	c.Assert(count, check.Equals, countBefore+1)
	c.Assert(successCount, check.Equals, countSuccessBefore+1)
}
