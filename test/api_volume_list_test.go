package main

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIVolumeListSuite is the test suite for volume inspect API.
type APIVolumeListSuite struct{}

func init() {
	check.Suite(&APIVolumeListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIVolumeListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestVolumeListOk tests if list volumes is OK.
func (suite *APIVolumeListSuite) TestVolumeListOk(c *check.C) {
	// Create a volume with the name "TestVolume1".
	err := CreateVolume(c, "TestVolume1", "local")
	c.Assert(err, check.IsNil)
	defer RemoveVolume(c, "TestVolume1")

	// Create a volume with the name "TestVolume1".
	err = CreateVolume(c, "TestVolume2", "local")
	c.Assert(err, check.IsNil)
	defer RemoveVolume(c, "TestVolume2")

	// Test volume list feature.
	path := "/volumes"
	resp, err := request.Get(path)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	// Check list result.
	volumeListResp := &types.VolumeListResp{}
	err = request.DecodeBody(volumeListResp, resp.Body)
	c.Assert(err, check.IsNil)

	// Check response having the pre-created two volumes.
	found := 0
	for _, volume := range volumeListResp.Volumes {
		if volume.Name == "TestVolume1" || volume.Name == "TestVolume2" {
			found++
		}
	}
	c.Assert(found, check.Equals, 2)
}
