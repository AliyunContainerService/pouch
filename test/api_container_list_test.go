package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIContainerListSuite is the test suite for container list API.
type APIContainerListSuite struct{}

func init() {
	check.Suite(&APIContainerListSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerListSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)
	PullImage(c, busyboxImage125)
}

// TestListAll test api with -all parameters, empty filter.
func (suite *APIContainerListSuite) TestListAll(c *check.C) {
	containerA := "TestListAllContainerA"
	resA := command.PouchRun("run", "-d", "--name", containerA, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerA)
	containerAID := strings.TrimSpace(resA.Combined())

	containerB := "TestListAllContainerB"
	command.PouchRun("create", "--name", containerB, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerB)

	// all: false
	success, got, _ := getContainerListOK(c, "", false)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	// check fields
	c.Assert(got[0].ID, check.Equals, containerAID)
	c.Assert(got[0].Image, check.Equals, busyboxImage125)
	c.Assert(got[0].ImageID, check.Equals, busyboxImage125ID)
	c.Assert(len(got[0].Names), check.Equals, 1)
	c.Assert(got[0].Names[0], check.Equals, containerA)

	// all: true
	success, got, _ = getContainerListOK(c, "", true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 2)
}

// TestListFilterInvalid test invalid filter.
func (suite *APIContainerListSuite) TestListFilterInvalid(c *check.C) {
	success, _, errResp := getContainerListOK(c, "foo", false)
	c.Assert(success, check.Equals, false)
	err := util.PartialEqual(errResp.Message, "invalid character")
	c.Assert(err, check.IsNil)

	success, _, errResp = getContainerListOK(c, "{\"foo\":[\"bar\"]}", false)
	c.Assert(success, check.Equals, false)
	err = util.PartialEqual(errResp.Message, "invalid filter")
	c.Assert(err, check.IsNil)

	success, _, errResp = getContainerListOK(c, "{\"id\":[\"null\"],\"foo\":[\"bar\"]}", false)
	c.Assert(success, check.Equals, false)
	err = util.PartialEqual(errResp.Message, "invalid filter")
	c.Assert(err, check.IsNil)
}

// TestListFilterInvalid test equal filter.
func (suite *APIContainerListSuite) TestListFilterEqual(c *check.C) {
	containerA := "TestListFilterEqualContainerA"
	resA := command.PouchRun("run", "-d", "--name", containerA, "-l", "label="+containerA, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerA)
	containerAID := strings.TrimSpace(resA.Combined())

	containerB := "TestListFilterEqualContainerB"
	resB := command.PouchRun("create", "--name", containerB, "-l", "label="+containerB, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerB)
	containerBID := strings.TrimSpace(resB.Combined())

	// id filter
	success, got, _ := getContainerListOK(c, fmt.Sprintf("{\"id\":[\"%s\"]}", containerAID), true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.Equals, containerAID)

	// name filter
	success, got, _ = getContainerListOK(c, fmt.Sprintf("{\"name\":[\"%s\"]}", containerA), true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.Equals, containerAID)

	// label filter
	success, got, _ = getContainerListOK(c, fmt.Sprintf("{\"label\":[\"label=%s\"]}", containerB), true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.Equals, containerBID)

	// status filter
	success, got, _ = getContainerListOK(c, fmt.Sprintf("{\"status\":[\"%s\"]}", "created"), true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.Equals, containerBID)

	// combined filter matched
	success, got, _ = getContainerListOK(c,
		fmt.Sprintf("{\"id\":[\"%s\"],\"status\":[\"%s\"],\"label\":[\"label=%s\"],\"name\":[\"%s\"]}", containerBID, "created", containerB, containerB),
		true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.Equals, containerBID)

	// combined filter
	success, got, _ = getContainerListOK(c,
		fmt.Sprintf("{\"id\":[\"%s\"],\"status\":[\"%s\"],\"label\":[\"label=%s\"],\"name\":[\"%s\"]}", containerAID, "created", containerB, containerB),
		true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 0)

	// regex filter
	success, got, _ = getContainerListOK(c, "{\"name\":[\"A\"]}", true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.Equals, containerAID)
}

// TestListFilterUnEqual test label unequal filter.
func (suite *APIContainerListSuite) TestListFilterUnEqual(c *check.C) {
	containerA := "TestListFilterUnEqualContainerA"
	resA := command.PouchRun("run", "-d", "-l", "label="+containerA, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerA)
	containerAID := strings.TrimSpace(resA.Combined())

	containerB := "TestListFilterUnEqualContainerB"
	resB := command.PouchRun("run", "-d", "-l", "label="+containerB, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerB)
	containerBID := strings.TrimSpace(resB.Combined())

	success, got, _ := getContainerListOK(c, fmt.Sprintf("{\"label\":[\"label!=%s\"]}", containerB), true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.Equals, containerAID)

	success, got, _ = getContainerListOK(c, fmt.Sprintf("{\"id\":[\"%s\"],\"label\":[\"label!=null\"]}", containerBID), true)
	c.Assert(success, check.Equals, true)
	c.Assert(len(got), check.Equals, 1)
	c.Assert(got[0].ID, check.Equals, containerBID)
}

func getContainerListOK(c *check.C, filters string, all bool) (success bool, got []types.Container, errResp types.Error) {
	q := url.Values{}
	q.Set("filters", filters)
	if all {
		q.Set("all", "true")
	}

	resp, err := request.Get("/containers/json", request.WithQuery(q))
	c.Assert(err, check.IsNil)

	if resp.StatusCode/100 == 2 {
		success = true
		err = request.DecodeBody(&got, resp.Body)
		c.Assert(err, check.IsNil)
	} else {
		success = false
		err = request.DecodeBody(&errResp, resp.Body)
		c.Assert(err, check.IsNil)
	}

	return success, got, errResp
}
