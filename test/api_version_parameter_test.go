package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// APIVersionSuite is the test suit for all about pouch API Version.
type APIVersionSuite struct{}

func init() {
	check.Suite(&APIVersionSuite{})
}

// SetupTest does common setup in the beginning of each test.
func (suite *APIVersionSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

// TestNoVersionParamsInURL test api url not contains version info.
// Pouch api url support with or without version info.
func (suite *APIVersionSuite) TestNoVersionParamsInURL(c *check.C) {
	cname := "TestCreateURLNoVersionInfo"

	commonAPIClient, err := client.NewAPIClient(environment.PouchdAddress, environment.TLSConfig)
	c.Assert(err, check.IsNil)
	apiClient := commonAPIClient.(*client.APIClient)

	// set version to "", let request url not contains version info.
	apiClient.UpdateClientVersion("")

	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	b, err := json.Marshal(obj)
	c.Assert(err, check.IsNil)
	body := bytes.NewReader(b)

	fullPath := apiClient.BaseURL() + apiClient.GetAPIPath("/containers/create", q)

	req, err := http.NewRequest("POST", fullPath, body)
	c.Assert(err, check.IsNil)

	resp, err := apiClient.HTTPCli.Do(req)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 201)
}
