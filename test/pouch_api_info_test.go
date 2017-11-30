package main

import (
	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"
	"net/http"

	"github.com/go-check/check"
)

// PouchAPIInfoSuite is the test suite for info related API.
type PouchAPIInfoSuite struct {
}

func init() {
	check.Suite(&PouchAPIInfoSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIInfoSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestInfo is a demo of API test.
func (suite *PouchAPIInfoSuite) TestInfo(c *check.C) {
	host := ""
	client, err := client.NewAPIClient(host, utils.TLSConfig{})
	c.Assert(err, check.IsNil)

	req, err := http.NewRequest("GET", client.BaseURL()+"/info", nil)
	c.Assert(err, check.IsNil)

	resp, err := client.HTTPCli.Do(req)
	c.Assert(err, check.IsNil)

	c.Assert(resp.StatusCode, check.Equals, 200)
}
