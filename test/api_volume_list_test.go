package main

import (
	"net/url"

	"github.com/alibaba/pouch/apis/filters"
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
	CreateVolumeOK(c, "TestVolume1", "local", nil)
	defer RemoveVolumeOK(c, "TestVolume1")

	// Create a volume with the name "TestVolume2".
	CreateVolumeOK(c, "TestVolume2", "local", nil)
	defer RemoveVolumeOK(c, "TestVolume2")

	// Create a volume with the name "TestVolume3".
	options := map[string]string{"mountpoint": "/data/TestVolume3"}
	CreateVolumeOK(c, "TestVolume3", "local", options)
	defer RemoveVolumeOK(c, "TestVolume3")

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
		if volume.Name == "TestVolume1" ||
			volume.Name == "TestVolume2" ||
			volume.Name == "TestVolume3" {
			found++
		}
	}
	c.Assert(found, check.Equals, 3)
}

// TestVolumeListFilter tests if list volumes with filter is OK.
func (suite *APIVolumeListSuite) TestVolumeListFilter(c *check.C) {
	// Create a volume with the name "TestVolume1".
	testVolumeName := "TestVolume1"
	CreateVolumeOK(c, testVolumeName, "local", nil)
	defer RemoveVolumeOK(c, testVolumeName)

	f := filters.NewArgs()
	f.Add("name", "TestVolume1")
	filterJSON, err := filters.ToParam(f)
	c.Assert(err, check.IsNil)

	q := url.Values{}
	q.Add("filters", filterJSON)
	query := request.WithQuery(q)
	resp, err := request.Get("/volumes", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	// Check list result.
	volumeListResp := &types.VolumeListResp{}
	err = request.DecodeBody(volumeListResp, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(len(volumeListResp.Volumes), check.Equals, 1)
	c.Assert(volumeListResp.Volumes[0].Name, check.Equals, testVolumeName)
}

// TestVolumeListInvalidFilter tests if list volumes with invalid filter is OK.
func (suite *APIVolumeListSuite) TestVolumeListInvalidFilter(c *check.C) {
	// Create a volume with the name "TestVolume1".
	testVolumeName := "TestVolume1"
	CreateVolumeOK(c, testVolumeName, "local", nil)
	defer RemoveVolumeOK(c, testVolumeName)

	f := filters.NewArgs()
	f.Add("driver-name", "TestVolume1")
	filterJSON, err := filters.ToParam(f)
	c.Assert(err, check.IsNil)

	q := url.Values{}
	q.Add("filters", filterJSON)
	query := request.WithQuery(q)
	resp, err := request.Get("/volumes", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 500)
}
