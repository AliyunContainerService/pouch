package main

import (
	"regexp"

	"github.com/alibaba/pouch/pkg/kernel"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/version"

	"github.com/go-check/check"
)

// APIEngineVersionMetricsSuite is the test suite for EngineVersion metrics API.
type APIEngineVersionMetricsSuite struct{}

func init() {
	check.Suite(&APIEngineVersionMetricsSuite{})
}

// SetUpSuite does common setup in the beginning of each suite.
func (suite *APIEngineVersionMetricsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestEngineVersionMetrics tests metrics of EngineVersion.
func (suite *APIEngineVersionMetricsSuite) TestEngineVersionMetrics(c *check.C) {
	commitExpected := version.GitCommit
	versionExpected := version.Version
	kernelVersion, err := kernel.GetKernelVersion()
	c.Assert(err, check.IsNil)
	kernelExpected := kernelVersion.String()

	key := "engine_daemon_engine_info{"
	versionMetrics := GetMetricLine(c, key)

	regularCommit := `^.*commit="(.*?)".*$`
	regular := regexp.MustCompile(regularCommit)
	params := regular.FindStringSubmatch(versionMetrics)
	c.Assert(params[1], check.Equals, commitExpected)

	regularVersion := `^.*version="(.*?)".*$`
	regular = regexp.MustCompile(regularVersion)
	params = regular.FindStringSubmatch(versionMetrics)
	c.Assert(params[1], check.Equals, versionExpected)

	regularKernel := `^.*kernel="(.*?)".*$`
	regular = regexp.MustCompile(regularKernel)
	params = regular.FindStringSubmatch(versionMetrics)
	c.Assert(params[1], check.Equals, kernelExpected)

}
