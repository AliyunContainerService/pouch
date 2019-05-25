package main

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIVolumePruneSuite is the test suite for volume prune API.
type APIVolumePruneSuite struct{}

func init() {
	check.Suite(&APIVolumePruneSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIVolumePruneSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestVolumeListOk tests if list volumes is OK.
func (suite *APIVolumePruneSuite) TestVolumePruneOk(c *check.C) {
	// Create a volume with the name "TestVolume1".
	CreateVolumeOK(c, "TestVolume1", "local", nil)

	// Create a volume with the name "TestVolume2".
	CreateVolumeOK(c, "TestVolume2", "local", nil)

	// Create a volume with the name "TestVolume3".
	options := map[string]string{"mountpoint": "/data/TestVolume3"}
	CreateVolumeOK(c, "TestVolume3", "local", options)

	// Test volume prune feature.
	resp, err := request.Post("/volumes/prune")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	// Check prune result.
	volumePruneResp := &types.VolumePruneResp{}
	err = request.DecodeBody(volumePruneResp, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(len(volumePruneResp.VolumesDeleted), check.Equals, 3)
}
