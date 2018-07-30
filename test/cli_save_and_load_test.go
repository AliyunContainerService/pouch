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

// TestSaveLoadWorks tests "pouch save" and "pouch load" work.
func (suite *PouchSaveLoadSuite) TestSaveLoadWorks(c *check.C) {
	res := command.PouchRun("pull", helloworldImage)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("image", "inspect", helloworldImage)
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

	res = command.PouchRun("save", "-o", filepath.Join(dir, "helloworld.tar"), helloworldImage)
	res.Assert(c, icmd.Success)

	loadImageName := "load-helloworld"
	res = command.PouchRun("load", "-i", filepath.Join(dir, "helloworld.tar"), loadImageName)
	res.Assert(c, icmd.Success)
	defer command.PouchRun("rmi", loadImageName+":"+environment.HelloworldTag)

	res = command.PouchRun("image", "inspect", loadImageName+":"+environment.HelloworldTag)
	res.Assert(c, icmd.Success)

	after := []types.ImageInfo{}
	if err := json.Unmarshal([]byte(res.Stdout()), &after); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(before[0].ID, check.Equals, after[0].ID)
	c.Assert(before[0].CreatedAt, check.Equals, after[0].CreatedAt)
	c.Assert(before[0].Size, check.Equals, after[0].Size)
}
