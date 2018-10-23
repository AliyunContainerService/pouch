package main

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchCreateLogOptionsSuite is the test suite for create CLI with LogOptions.
type PouchCreateLogOptionsSuite struct{}

func init() {
	check.Suite(&PouchCreateLogOptionsSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchCreateLogOptionsSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestFailNoneWithOpts fails.
func (suite *PouchCreateLogOptionsSuite) TestFailNoneWithOpts(c *check.C) {
	cname := "TestCreateLogOptions_Fail_none_with_opts"
	expected := "don't allow to set logging opts for driver none"

	args := []string{"create"}
	args = append(args, getArgsForLogOptions("none", []string{"tag=1"})...)
	args = append(args, "--name", cname, busyboxImage)

	res := command.PouchRun(args...)
	if got := res.Combined(); !strings.Contains(got, expected) {
		c.Fatalf("expected to contains (%v), but got (%v)", expected, got)
	}
}

// TestFailNotSupportDriver fails.
func (suite *PouchCreateLogOptionsSuite) TestFailNotSupportDriver(c *check.C) {
	cname := "TestCreateLogOptions_Fail_not_support_driver"

	driver := "notyet"
	expected := "validation failure"

	args := []string{"create"}
	args = append(args, getArgsForLogOptions(driver, nil)...)
	args = append(args, "--name", cname, busyboxImage)

	res := command.PouchRun(args...)
	if got := res.Combined(); !strings.Contains(got, expected) {
		c.Fatalf("expected to contains (%v), but got (%v)", expected, got)
	}
}

// TestOK tests happy cases for log options
func (suite *PouchCreateLogOptionsSuite) TestOK(c *check.C) {
	type tCase struct {
		cname    string
		driver   string
		logOpts  []string
		expected map[string]string
	}

	for _, tc := range []tCase{
		{
			cname:    "TestCreateLogOptions_none",
			driver:   "none",
			expected: nil,
		}, {
			cname:    "TestCreateLogOptions_jsonfile",
			driver:   "json-file",
			expected: nil,
		}, {
			cname:   "TestCreateLogOptions_jsonfile_tag=1",
			driver:  "json-file",
			logOpts: []string{"tag=1"},
			expected: map[string]string{
				"tag": "1",
			},
		},
	} {
		args := []string{"create"}
		args = append(args, getArgsForLogOptions(tc.driver, tc.logOpts)...)
		args = append(args, "--name", tc.cname, busyboxImage)

		command.PouchRun(args...).Assert(c, icmd.Success)
		defer DelContainerForceMultyTime(c, tc.cname)

		cfg := suite.getContainerLogConfig(c, tc.cname)
		c.Assert(cfg.LogDriver, check.Equals, tc.driver)

		if !reflect.DeepEqual(cfg.LogOpts, tc.expected) {
			c.Errorf("expected to have (%v), but got (%v)", tc.expected, cfg.LogOpts)
		}
	}
}

func getArgsForLogOptions(driver string, logOpts []string) []string {
	args := []string{"--log-driver", driver}
	for _, opt := range logOpts {
		args = append(args, "--log-opt", opt)
	}
	return args
}

func (suite *PouchCreateLogOptionsSuite) getContainerLogConfig(c *check.C, idOrName string) *types.LogConfig {
	output := command.PouchRun("inspect", idOrName).Combined()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(result, check.HasLen, 1)
	return result[0].HostConfig.LogConfig
}
