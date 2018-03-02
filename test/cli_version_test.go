package main

import (
	"regexp"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/version"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchVersionSuite is the test suite for version CLI.
type PouchVersionSuite struct{}

func init() {
	check.Suite(&PouchVersionSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchVersionSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestPouchVersion is to verify pouch version.
//
// TODO: check more values if version functionality is ready.
func (suite *PouchVersionSuite) TestPouchVersion(c *check.C) {
	res := command.PouchRun("version").Assert(c, icmd.Success)
	kv := versionToKV(res.Combined())

	c.Assert(kv["GoVersion"], check.Equals, version.GOVersion)
	c.Assert(kv["APIVersion"], check.Equals, version.APIVersion)
	c.Assert(kv["Arch"], check.Equals, runtime.GOARCH)
	c.Assert(kv["Os"], check.Equals, runtime.GOOS)
	c.Assert(kv["Version"], check.Equals, version.Version)
}

// versionToKV reads version string into key-value mapping.
func versionToKV(version string) map[string]string {
	res := make(map[string]string)

	reg := regexp.MustCompile(`^\w+:`)

	lines := strings.Split(version, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if loc := reg.FindStringIndex(line); loc != nil {
			k, v := line[:loc[1]-1], line[loc[1]:]
			res[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return res
}
