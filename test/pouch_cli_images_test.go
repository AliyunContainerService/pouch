package main

import (
	"github.com/go-check/check"
	"os/exec"
	"strings"
)

// PouchImagesSuite is the test suite of help CLI.
type PouchImagesSuite struct {
}

func init() {
	check.Suite(&PouchImagesSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchImagesSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchImages is to verify pouch images.
func (suite *PouchImagesSuite) TestPouchImages(c *check.C) {
	out, err := exec.Command("pouch", "images").Output()
	c.Assert(err, check.IsNil)

	if !strings.Contains(string(out), "IMAGE") {
		c.Fatalf("unexpected output %s expected IMAGE\n", string(out))
	}
}

// TestPouchImagesFlags is to verify pouch images with flags
func (suite *PouchImagesSuite) TestPouchImagesFlags(c *check.C) {
	args := map[string]bool{
		"-q":       true,
		"--digest": true,
	}

	for arg, ok := range args {
		cmd := exec.Command("pouch", "images", arg)
		_, _, err := runCmd(cmd)

		if ok {
			c.Assert(err, check.IsNil)
		} else {
			c.Assert(err, check.NotNil)
		}
	}
}
