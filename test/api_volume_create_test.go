package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIVolumeCreateSuite is the test suite for volume create API.
type APIVolumeCreateSuite struct{}

func init() {
	check.Suite(&APIVolumeCreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIVolumeCreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestVolumeCreateOk tests creating a volume is OK.
func (suite *APIVolumeCreateSuite) TestVolumeCreateOk(c *check.C) {
	vol := "TestVolumeCreateOk"

	CreateVolumeOK(c, vol, "local", nil)
	RemoveVolumeOK(c, vol)
}

// TestPluginVolumeCreateOk tests creating a volume which created by volume plugin
func (suite *APIVolumeCreateSuite) TestPluginVolumeCreateOk(c *check.C) {
	vol := "TestPluginVolumeCreateOk"

	CreateVolumeOK(c, vol, "local-persist", map[string]string{"mountpoint": "/data/images"})
	RemoveVolumeOK(c, vol)
}
