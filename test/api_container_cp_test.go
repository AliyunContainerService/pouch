package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIContainerCopySuite is the test suite for container cp API.
type APIContainerCopySuite struct{}

func init() {
	check.Suite(&APIContainerCopySuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerCopySuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// TestCopyWorks test pouch cp works well
func (suite *APIContainerCopySuite) TestCopyWorks(c *check.C) {
	cname := "TestCopyWorks"
	command.PouchRun("run", "--name", cname, busyboxImage,
		"sh", "-c",
		"echo 'test pouch cp' >> data.txt").Assert(c, icmd.Success)

	defer DelContainerForceMultyTime(c, cname)
	dstPath := "data.txt"

	// stats file in container
	q := url.Values{}
	q.Set("path", dstPath)
	query := request.WithQuery(q)
	resp, err := request.Head("/containers/"+cname+"/archive", query)
	c.Check(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	stat, err := getContainerPathStatFromHeader(resp.Header)
	c.Assert(err, check.IsNil)
	c.Assert(stat.Mode, check.NotNil)
	c.Assert(stat.Mtime, check.NotNil)
	c.Assert(strings.Contains(stat.Name, dstPath), check.Equals, true)
	c.Assert(strings.Contains(stat.Path, dstPath), check.Equals, true)

	// test copy file from container
	q = url.Values{}
	q.Set("path", dstPath)
	query = request.WithQuery(q)

	resp, err = request.Get("/containers/"+cname+"/archive", query)
	c.Check(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	c.Assert(len(body) > 0, check.Equals, true)

	//TODO: add test case copy file to container
}

func getContainerPathStatFromHeader(header http.Header) (types.ContainerPathStat, error) {
	var stat types.ContainerPathStat
	encodedStat := header.Get("X-Docker-Container-Path-Stat")
	statDecoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encodedStat))

	err := json.NewDecoder(statDecoder).Decode(&stat)
	if err != nil {
		err = fmt.Errorf("unable to decode container path stat header: %s", err)
	}
	return stat, err
}
