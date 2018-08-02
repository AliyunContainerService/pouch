package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIVolumeInspectSuite is the test suite for volume inspect API.
type APIVolumeInspectSuite struct{}

func init() {
	check.Suite(&APIVolumeInspectSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIVolumeInspectSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestVolumeInspectOk tests if inspecting a volume is OK.
func (suite *APIVolumeInspectSuite) TestVolumeInspectOk(c *check.C) {
	// Create a volume with the name "TestVolume".
	vol := "TestVolume"
	CreateVolumeOK(c, vol, "local", nil)
	defer RemoveVolumeOK(c, vol)

	// Test volume inspect feature.
	path := "/volumes/" + vol
	resp, err := request.Get(path)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	// TODO: missing case
	//
	//	add field check
}

// TestVolumeInspectNotFound tests if inspecting a nonexistent volume returns error.
func (suite *APIVolumeInspectSuite) TestVolumeInspectNotFound(c *check.C) {
	vol := "NotExistentVolume"

	path := "/volumes/" + vol
	resp, err := request.Get(path)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}
