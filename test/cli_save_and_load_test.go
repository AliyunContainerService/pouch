package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchSaveLoadSuite is the test suite for save and load CLI.
type PouchSaveLoadSuite struct{}

func init() {
	check.Suite(&PouchSaveLoadSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchSaveLoadSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	environment.PruneAllContainers(apiClient)
}

// TestSaveLoadDockerImages tests "pouch load" docker images.
func (suite *PouchSaveLoadSuite) TestSaveLoadDockerImages(c *check.C) {
	environment.PruneAllImages(apiClient)

	// the tar file contains the busybox:1.25 and alpine:3.7
	filename := filepath.Join("testdata", "images", "docker-busybox_1_25-and-alpine_3_7.tar")
	command.PouchRun("load", "-i", filename).Assert(c, icmd.Success)

	command.PouchRun("image", "inspect", "docker.io/library/busybox:1.25").Assert(c, icmd.Success)
	command.PouchRun("image", "inspect", "docker.io/library/alpine:3.7").Assert(c, icmd.Success)
}

// TestSaveLoadOneDockerImage tests "pouch load -i <docker images> one-image.
func (suite *PouchSaveLoadSuite) TestSaveLoadOneDockerImage(c *check.C) {
	environment.PruneAllImages(apiClient)

	// the tar file contains the busybox:1.25 and alpine:3.7
	filename := filepath.Join("testdata", "images", "docker-busybox_1_25-and-alpine_3_7.tar")

	// only load alpine
	command.PouchRun("load", "-i", filename, "docker.io/library/alpine").Assert(c, icmd.Success)

	command.PouchRun("image", "inspect", "docker.io/library/alpine:3.7").Assert(c, icmd.Success)

	// busybox should be ignored
	res := command.PouchRun("image", "inspect", "docker.io/library/busybox:1.25")
	c.Assert(res.ExitCode, check.Not(check.Equals), 0)
}

// TestSaveLoadWorks tests "pouch save" and "pouch load" work.
func (suite *PouchSaveLoadSuite) TestSaveLoadWorks(c *check.C) {
	res := command.PouchRun("pull", busyboxImage125)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("image", "inspect", busyboxImage125)
	res.Assert(c, icmd.Success)

	before := []types.ImageInfo{}
	if err := json.Unmarshal([]byte(res.Stdout()), &before); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	dir, err := ioutil.TempDir("", "TestSaveLoadWorks")
	if err != nil {
		c.Errorf("failed to create a new temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	res = command.PouchRun("save", "-o", filepath.Join(dir, "busyboxImage.tar"), busyboxImage125)
	res.Assert(c, icmd.Success)

	loadImageName := "load-busyboxImage"
	res = command.PouchRun("load", "-i", filepath.Join(dir, "busyboxImage.tar"), loadImageName)
	res.Assert(c, icmd.Success)
	defer command.PouchRun("rmi", loadImageName+":"+environment.Busybox125Tag)

	res = command.PouchRun("image", "inspect", loadImageName+":"+environment.Busybox125Tag)
	res.Assert(c, icmd.Success)

	after := []types.ImageInfo{}
	if err := json.Unmarshal([]byte(res.Stdout()), &after); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(before[0].ID, check.Equals, after[0].ID)
	c.Assert(before[0].CreatedAt, check.Equals, after[0].CreatedAt)
	c.Assert(before[0].Size, check.Equals, after[0].Size)
}
