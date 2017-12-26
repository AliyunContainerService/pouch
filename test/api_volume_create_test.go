package main

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

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

	obj := map[string]interface{}{
		"Driver": "local",
		"Name":   vol,
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post(c, "/volumes/create", body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())

	resp, err = request.Delete(c, "/volumes/"+vol)
	c.Assert(resp.StatusCode, check.Equals, 200, err.Error())
}
