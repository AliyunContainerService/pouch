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

	cmd := PouchCmd{
		args:   []string{"volume", "create", "--name", funcname},
		result: true,
	}
	RunCmd(c, &cmd)

	cmd = PouchCmd{
		args:   []string{"volume", "remove", funcname},
		result: true,
	}
	RunCmd(c, &cmd)

}
