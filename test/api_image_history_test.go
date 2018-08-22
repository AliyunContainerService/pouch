package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImageHistorySuite is the test suite for image history API.
type APIImageHistorySuite struct{}

func init() {
	check.Suite(&APIImageHistorySuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageHistorySuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestImageHistoryOk tests getting image history is OK.
func (suite *APIImageHistorySuite) TestImageHistoryOk(c *check.C) {
	// TODO: We shouldn't compare dockerhub's image history with a fixed string, that's too unreadable.
	// So this test case will be done when pouch enables build functionality.
}

// TestImageHistoryNotFound tests getting history of non-existing image return 404.
func (suite *APIImageHistorySuite) TestImageHistoryNotFound(c *check.C) {
	img := "TestImageHistoryNotFound"
	resp, err := request.Get("/images/" + img + "/history")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}
