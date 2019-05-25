package main

import (
	"fmt"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImageMetricsSuite is the test suite for image metrics API.
type APIImageMetricsSuite struct{}

func init() {
	check.Suite(&APIImageMetricsSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageMetricsSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// SetUpSuite does common setup in the beginning of each suite .
func (suite *APIImageMetricsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, helloworldImage)
	DelImageForceOk(c, helloworldImage)
}

// TearDownSuite run after each suite to do cleanup work for the whole suite.
func (suite *APIImageMetricsSuite) TearDownSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestDeleteImageOk tests metrics of image.
func (suite *APIImageMetricsSuite) TestImageMetrics(c *check.C) {
	suite.checkAction(c, "pull")
	suite.checkAction(c, "delete")
}

func (suite *APIImageMetricsSuite) checkAction(c *check.C, label string) {
	key := fmt.Sprintf(`engine_daemon_image_actions_total{action="%s"}`, label)
	keySuccess := fmt.Sprintf(`engine_daemon_image_success_actions_total{action="%s"}`, label)
	countBefore, countSuccessBefore := GetMetric(c, key, keySuccess)

	switch label {
	case "pull":
		PullImage(c, helloworldImage)
	case "delete":
		resp, err := request.Delete("/images/" + helloworldImage)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 204)
	}

	count, successCount := GetMetric(c,
		key,
		keySuccess)
	c.Assert(count, check.Equals, countBefore+1)
	c.Assert(successCount, check.Equals, countSuccessBefore+1)
}
