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

	path := "/volumes/create"
	body := request.WithJSONBody(obj)
	resp, err := request.Post(path, body)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 201)

	path = "/volumes/" + vol
	resp, err = request.Delete(path)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, 200)
}
