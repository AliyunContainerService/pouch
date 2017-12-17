package main

import (
	"os/exec"

	"github.com/go-check/check"
)

// PouchUnpauseSuite is the test suite fo help CLI.
type PouchUnpauseSuite struct {
}

func init() {
	check.Suite(&PouchUnpauseSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchUnpauseSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

	// Pull test image
	cmd := exec.Command("pouch", "pull", testImage)
	cmd.Run()
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchUnpauseSuite) SetUpTest(c *check.C) {
	// TODO
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchUnpauseSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchUnpauseSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestStopWorks tests "pouch unpause" work.
func (suite *PouchUnpauseSuite) TestUnpauseWorks(c *check.C) {
	containernames := []string{"bar1", "bar2"}
	for _, containername := range containernames {
		cmd := PouchCmd{
			args:   []string{"create", "--name", containername, testImage},
			result: true,
		}
		RunCmd(c, &cmd)

		cmd = PouchCmd{
			args:   []string{"start", containername},
			result: true,
		}
		RunCmd(c, &cmd)
	}

	cmd := PouchCmd{
		args:   []string{"pause", containernames[0]},
		result: true,
	}
	RunCmd(c, &cmd)

	args := map[string]bool{
		// paused container
		containernames[0]: true,
		// running container
		containernames[1]: false,
		// not exist
		"bar3": false,
	}

	for arg, ok := range args {
		cmd := PouchCmd{
			args:   []string{"unpause", arg},
			result: ok,
		}
		RunCmd(c, &cmd)
	}
}
