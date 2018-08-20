package main

import (
	"net/url"

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

	DelContainerForceMultyTime(c, newname)
}

// TestRenameOk test create api is ok with default parameters.
func (suite *APIContainerRenameSuite) TestRenameById(c *check.C) {
	oldname := "TestRenameOk"
	newname := "NewTestRenameOk"

	// create container by oldname and change it into newname by container id
	{
		cid := CreateBusyboxContainerOk(c, oldname, "top")
		newq := url.Values{}
		newq.Add("name", newname)
		resp, err := request.Post("/containers/"+cid+"/rename", request.WithQuery(newq))
		c.Assert(err, check.IsNil)
		defer resp.Body.Close()
		CheckRespStatus(c, resp, 204)

		defer DelContainerForceMultyTime(c, newname)
	}

	// after oldname release, the oldname is available so that create should OK
	{
		CreateBusyboxContainerOk(c, oldname, "top")
		DelContainerForceMultyTime(c, oldname)
	}
}
