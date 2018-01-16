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
	obj := map[string]interface{}{
		"Driver": "local",
		"Name":   vol,
	}
	path := "/volumes/create"
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	// Test volume inspect feature.
	path = "/volumes/" + vol
	resp, err = request.Get(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)

	// Delete the volume.
	path = "/volumes/" + vol
	resp, err = request.Delete(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 204)
}

// TestVolumeInspectNotFound tests if inspecting a nonexistent volume returns error.
func (suite *APIVolumeInspectSuite) TestVolumeInspectNotFound(c *check.C) {
	vol := "NotExistentVolume"

	path := "/volumes/" + vol
	resp, err := request.Get(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 404)

}
