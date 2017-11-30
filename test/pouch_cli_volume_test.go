package main

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/go-check/check"
)

// PouchVolumeSuite is the test suite fo help CLI.
type PouchVolumeSuite struct {
}

func init() {
	check.Suite(&PouchVolumeSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchVolumeSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

	// Pull test image
	cmd := exec.Command("pouch", "pull", testImage)
	cmd.Run()
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchVolumeSuite) SetUpTest(c *check.C) {
	// TODO
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchVolumeSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchVolumeSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestVolumeWorks tests "pouch volume" work.
func (suite *PouchVolumeSuite) TestVolumeWorks(c *check.C) {

	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	cmd := exec.Command("pouch", "volume", "create", "--name", funcname)
	runCmdPos(c, cmd)

	cmd = exec.Command("pouch", "volume", "remove", funcname)
	runCmdPos(c, cmd)
}
