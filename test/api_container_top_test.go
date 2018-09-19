package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerTopSuite is the test suite for container top API.
type APIContainerTopSuite struct{}

func init() {
	check.Suite(&APIContainerTopSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerTopSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestTopContainer is to verify the correctness of pouch top command.
func (suite *APIContainerTopSuite) TestTopContainer(c *check.C) {
	cname := "TestTop"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	resp, err := request.Get("/containers/" + cname + "/top")
	c.Assert(err, check.IsNil)

	response := types.ContainerProcessList{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if resp != nil && resp.Body != nil {
		// close body ReadCloser to make Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}

	c.Assert(err, check.IsNil)
	c.Assert(response, check.NotNil)
	c.Assert(response.Titles[0], check.Equals, "UID")

	if len(response.Processes) != 1 {
		c.Fatalf("unexpected processes length %d expected %d", len(response.Processes), 1)
	}
}

// TestTopContainerWithOptions is to verify the correctness of pouch top command with ps options.
func (suite *APIContainerTopSuite) TestTopContainerWithOptions(c *check.C) {
	cname := "TestTopWithOptions"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)

	q := url.Values{}
	q.Add("ps_args", "-aux")
	query := request.WithQuery(q)
	resp, err := request.Get("/containers/"+cname+"/top", query)
	c.Assert(err, check.IsNil)

	response := types.ContainerProcessList{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if resp != nil && resp.Body != nil {
		// close body ReadCloser to make Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}
	c.Assert(err, check.IsNil)
	c.Assert(response, check.NotNil)
	c.Assert(response.Titles[0], check.Equals, "USER")

	if len(response.Processes) != 1 {
		c.Fatalf("unexpected processes length %d expected %d", len(response.Processes), 1)
	}
}
