package main

import (
	"os/exec"
	"regexp"

	"github.com/go-check/check"
)

// PouchImagesSuite is the test suite fo help CLI.
type PouchImagesSuite struct {
}

func init() {
	check.Suite(&PouchImagesSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchImagesSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

	// Pull test image
	cmd := exec.Command("pouch", "pull", testImage)
	cmd.Run()
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchImagesSuite) SetUpTest(c *check.C) {
	// TODO
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchImagesSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchImagesSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestImagesQuietOption is to verify the quiet flag.
func (suite *PouchImagesSuite) TestImagesQuietFlag(c *check.C) {
	qOut, _, err := runCmd(exec.Command("pouch", "images", "-q"))
	c.Assert(err, check.IsNil)

	quietOut, _, err := runCmd(exec.Command("pouch", "images", "--quiet"))
	c.Assert(err, check.IsNil)

	c.Assert(qOut, check.Equals, quietOut)
	if match, _ := regexp.MatchString("^[0-9a-f]+\n$", qOut); !match {
		c.Fatalf("should return numeric ID, but got %s", qOut)
	}
}

// TestImagesWorks tests "pouch image" work.
func (suite *PouchImagesSuite) TestImagesWorks(c *check.C) {

	// TODO: nil input should return success
	// TODO: add wrong args.
	args := map[string]bool{
		"": false,
	}

	for arg, ok := range args {
		cmd := exec.Command("pouch", "images", arg)

		if ok {
			runCmdPos(c, cmd)
		} else {
			runCmdNeg(c, cmd)
		}
	}
}
