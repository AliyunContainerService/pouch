package main

import (
	"net/http"
	"time"

	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIContainerWaitSuite is the test suite for container wait API.
type APIContainerWaitSuite struct{}

func init() {
	check.Suite(&APIContainerWaitSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerWaitSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestWaitOk tests waiting a stopped container return 200.
func (suite *APIContainerWaitSuite) TestWaitOk(c *check.C) {
	cname := "TestWaitOk"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)

	StartContainerOk(c, cname)
	StopContainerOk(c, cname)

	resp, err := request.Post("/containers/" + cname + "/wait")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)
}

// TestWaitRunningContainer tests waiting a running container to stop, then returns 200.
func (suite *APIContainerWaitSuite) TestWaitRunningContainer(c *check.C) {
	cname := "TestWaitRunningContainer"

	CreateBusyboxContainerOk(c, cname)
	defer DelContainerForceMultyTime(c, cname)
	StartContainerOk(c, cname)

	var (
		err  error
		resp *http.Response
	)

	chWait := make(chan struct{})
	go func() {
		chWait <- struct{}{}
		resp, err = request.Post("/containers/" + cname + "/wait")
		close(chWait)
	}()
	<-chWait
	time.Sleep(100 * time.Millisecond)
	StopContainerOk(c, cname)

	select {
	case <-chWait:
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 200)
	case <-time.After(2 * time.Second):
		c.Errorf("timeout waiting for `pouch wait` API to exit")
	}
}

// TestWaitNonExistingContainer tests waiting a non-existing container return 404.
func (suite *APIContainerWaitSuite) TestWaitNonExistingContainer(c *check.C) {
	cname := "TestNonExistingContainer"
	resp, err := request.Post("/containers/" + cname + "/wait")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}
