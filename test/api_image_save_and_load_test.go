package main

import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// APIImageSaveLoadSuite is the test suite for image save and load API.
type APIImageSaveLoadSuite struct{}

func init() {
	check.Suite(&APIImageSaveLoadSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIImageSaveLoadSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage125)
}

// TestImageSaveLoadOk tests saving and loading images are OK.
func (suite *APIImageSaveLoadSuite) TestImageSaveLoadOk(c *check.C) {
	before, err := request.Get("/images/" + busyboxImage125 + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, before, 200)
	gotBefore := types.ImageInfo{}
	err = request.DecodeBody(&gotBefore, before.Body)
	c.Assert(err, check.IsNil)

	q := url.Values{}
	q.Set("name", busyboxImage125)
	query := request.WithQuery(q)
	resp, err := request.Get("/images/save", query)
	c.Assert(err, check.IsNil)
	defer resp.Body.Close()

	dir, err := ioutil.TempDir("", "TestImageSaveLoadOk")
	if err != nil {
		c.Errorf("failed to create a new temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	tmpFile := filepath.Join(dir, "busyboxImage.tar")
	f, err := os.Create(tmpFile)
	if err != nil {
		c.Errorf("failed to create file: %v", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		c.Errorf("failed to save data to file: %v", err)
	}

	data, err := os.Open(tmpFile)
	if err != nil {
		c.Errorf("failed to load file's data: %v", err)
	}

	loadImageName := "load-busyboxImage"
	q = url.Values{}
	q.Set("name", loadImageName)

	query = request.WithQuery(q)
	reader := request.WithRawData(data)
	header := request.WithHeader("Content-Type", "application/x-tar")

	resp, err = request.Post("/images/load", query, reader, header)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	after, err := request.Get("/images/" + loadImageName + ":" + environment.Busybox125Tag + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, after, 200)
	defer request.Delete("/images/" + loadImageName + ":" + environment.Busybox125Tag)

	gotAfter := types.ImageInfo{}
	err = request.DecodeBody(&gotAfter, after.Body)
	c.Assert(err, check.IsNil)

	c.Assert(gotBefore.ID, check.Equals, gotAfter.ID)
	c.Assert(gotBefore.CreatedAt, check.Equals, gotAfter.CreatedAt)
	c.Assert(gotBefore.Size, check.Equals, gotAfter.Size)
}
