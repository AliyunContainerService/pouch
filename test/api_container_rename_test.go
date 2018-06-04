package main

import (
	"encoding/json"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerRenameSuite is the test suite for container create API.
type APIContainerRenameSuite struct{}

func init() {
	check.Suite(&APIContainerRenameSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerRenameSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestRenameOk test create api is ok with default parameters.
func (suite *APIContainerRenameSuite) TestRenameOk(c *check.C) {

	// must required
	oldname := "TestRenameOk"
	newname := "NewTestRenameOk"

	CreateBusyboxContainerOk(c, oldname)

	newq := url.Values{}
	newq.Add("name", newname)
	resp, err := request.Post("/containers/"+oldname+"/rename", request.WithQuery(newq))
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	DelContainerForceOk(c, newname)
}

// TestRenameOk test create api is ok with default parameters.
func (suite *APIContainerRenameSuite) TestRenameById(c *check.C) {
	oldname := "TestRenameOk"
	newname := "NewTestRenameOk"

	resp, err := CreateBusyboxContainer(c, oldname, "top")
	c.Assert(err, check.IsNil)
	defer resp.Body.Close()

	ccr := types.ContainerCreateResp{}
	err = json.NewDecoder(resp.Body).Decode(&ccr)
	c.Assert(err, check.IsNil)
	cid := ccr.ID

	newq := url.Values{}
	newq.Add("name", newname)
	resp2, err := request.Post("/containers/"+cid+"/rename", request.WithQuery(newq))
	c.Assert(err, check.IsNil)
	defer resp2.Body.Close()
	CheckRespStatus(c, resp2, 204)

	DelContainerForceOk(c, newname)

	resp3, err := CreateBusyboxContainer(c, oldname, "top")
	c.Assert(err, check.IsNil)
	defer resp3.Body.Close()

	ccr = types.ContainerCreateResp{}
	err = json.NewDecoder(resp3.Body).Decode(&ccr)
	c.Assert(err, check.IsNil)

	DelContainerForceOk(c, oldname)

	if ccr.ID == "" {
		c.Errorf("container with old name %s create error. %v", oldname, ccr)
	}
}
