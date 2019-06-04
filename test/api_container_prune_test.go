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

// APIContainerPruneSuite is the test suite for container prune API.
type APIContainerPruneSuite struct{}

func init() {
	check.Suite(&APIContainerPruneSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerPruneSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)
	PullImage(c, busyboxImage125)
}

// TestPruneAll test api with -all parameters, empty filter.
func (suite *APIContainerPruneSuite) TestPruneAll(c *check.C) {
	containerA := "TestPruneAllContainerA"
	command.PouchRun("run", "-d", "--name", containerA, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, containerA)

	containerB := "TestPruneAllContainerB"
	command.PouchRun("run", "-d", "--name", containerB, busyboxImage125, "top").Assert(c, icmd.Success)
	command.PouchRun("stop", containerB).Assert(c, icmd.Success)

	containerC := "TestPruneAllContainerC"
	command.PouchRun("run", "-d", "--name", containerC, busyboxImage125, "top").Assert(c, icmd.Success)
	command.PouchRun("stop", containerC)

	success, got, _ := getContainerPruneRes(c, "")
	c.Assert(success, check.Equals, true)
	c.Assert(len(got.ContainersDeleted), check.Equals, 2)
}

// TestPruneFilterInvalid test invalid filter.
func (suite *APIContainerPruneSuite) TestPruneFilterInvalid(c *check.C) {
	success, _, errResp := getContainerPruneRes(c, "foo")
	c.Assert(success, check.Equals, false)
	err := util.PartialEqual(errResp.Message, "invalid character")
	c.Assert(err, check.IsNil)

	success, _, errResp = getContainerPruneRes(c, "{\"foo\":[\"bar\"]}")
	c.Assert(success, check.Equals, false)
	err = util.PartialEqual(errResp.Message, "invalid filter")
	c.Assert(err, check.IsNil)

	success, _, errResp = getContainerPruneRes(c, "{\"id\":[\"null\"],\"foo\":[\"bar\"]}")
	c.Assert(success, check.Equals, false)
	err = util.PartialEqual(errResp.Message, "invalid filter")
	c.Assert(err, check.IsNil)
}

// TestPruneFilterValid test valid filter.
func (suite *APIContainerPruneSuite) TestPruneFilterValid(c *check.C) {
	container1 := "TestPruneFilterValidContainer1"
	command.PouchRun("run", "-d", "--name", container1, "-l", "label="+container1, busyboxImage125, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, container1)

	container2 := "TestPruneFilterValidContainer2"
	res2 := command.PouchRun("run", "-d", "--name", container2, "-l", "label="+container2, busyboxImage125, "top").Assert(c, icmd.Success)
	command.PouchRun("stop", container2).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, container2)
	container2ID := strings.TrimSpace(res2.Combined())

	container3 := "TestPruneFilterValidContainer3"
	res3 := command.PouchRun("run", "-d", "--name", container3, "-l", "label="+container3, busyboxImage125, "top").Assert(c, icmd.Success)
	command.PouchRun("stop", container3).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, container3)
	container3ID := strings.TrimSpace(res3.Combined())

	container4 := "TestPruneFilterValidContainer4"
	res4 := command.PouchRun("run", "-d", "--name", container4, "-l", "label="+container4, busyboxImage125, "top").Assert(c, icmd.Success)
	command.PouchRun("stop", container4).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, container4)
	container4ID := strings.TrimSpace(res4.Combined())

	// id filter
	success, got, _ := getContainerPruneRes(c, fmt.Sprintf("{\"id\":[\"%s\"]}", container2ID))
	c.Assert(success, check.Equals, true)
	c.Assert(len(got.ContainersDeleted), check.Equals, 1)
	c.Assert(got.ContainersDeleted[0], check.Equals, container2ID)

	// name filter
	success, got, _ = getContainerPruneRes(c, fmt.Sprintf("{\"name\":[\"%s\"]}", container3))
	c.Assert(success, check.Equals, true)
	c.Assert(len(got.ContainersDeleted), check.Equals, 1)
	c.Assert(got.ContainersDeleted[0], check.Equals, container3ID)

	// label filter
	success, got, _ = getContainerPruneRes(c, fmt.Sprintf("{\"label\":[\"label=%s\"]}", container4))
	c.Assert(success, check.Equals, true)
	c.Assert(len(got.ContainersDeleted), check.Equals, 1)
	c.Assert(got.ContainersDeleted[0], check.Equals, container4ID)
}

func getContainerPruneRes(c *check.C, filters string) (success bool, got types.ContainerPruneResp, errResp types.Error) {
	q := url.Values{}
	q.Set("filters", filters)

	resp, err := request.Post("/containers/prune", request.WithQuery(q))
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
