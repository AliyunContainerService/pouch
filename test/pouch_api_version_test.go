package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/version"

	"github.com/go-check/check"
)

// PouchAPIVersionSuite is the test suite for version API.
type PouchAPIVersionSuite struct {
}

func init() {
	check.Suite(&PouchAPIVersionSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchAPIVersionSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestVersion is a demo of API test.
func (suite *PouchAPIVersionSuite) TestVersion(c *check.C) {
	host := ""
	client, err := client.NewAPIClient(host, utils.TLSConfig{})
	c.Assert(err, check.IsNil)

	req, err := http.NewRequest("GET", client.BaseURL()+"/version", nil)
	c.Assert(err, check.IsNil)

	// send raw http request
	resp, err := client.HTTPCli.Do(req)
	c.Assert(err, check.IsNil)
	defer resp.Body.Close()

	// checkout http status code
	c.Assert(resp.StatusCode, check.Equals, 200)

	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)

	var outputVersion types.SystemVersion
	err = json.Unmarshal(data, &outputVersion)
	c.Assert(err, check.IsNil)

	// check response content
	c.Assert(outputVersion.Version, check.Equals, version.Version)
	c.Assert(outputVersion.APIVersion, check.Equals, version.APIVersion)
	// TODO: check more response content
}
