package main

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIDaemonUpdateSuite is the test suite for daemon update API.
type APIDaemonUpdateSuite struct{}

func init() {
	check.Suite(&APIDaemonUpdateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIDaemonUpdateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestStartOk tests starting container could work.
func (suite *APIDaemonUpdateSuite) TestUpdateDaemon(c *check.C) {
	labels := []string{
		"storage=ssd",
		"zone=shanghai",
	}

	obj := map[string]interface{}{
		"Labels":     labels,
		"ImageProxy": "http://192.168.0.3:2378",
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/daemon/update", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	resp, err = request.Get("/info", body)

	info := types.SystemInfo{}
	err = request.DecodeBody(&info, resp.Body)
	c.Assert(err, check.IsNil)

	for _, label := range labels {
		isContained := false
		for _, infoLabel := range info.Labels {
			if infoLabel == label {
				isContained = true
				break
			}
		}
		if !isContained {
			c.Fatalf("label %s should be in labels in info API", label)
		}
	}
	//c.Assert(info.Labels, check.DeepEquals, labels)
	// TODO: add checking image proxy
}
