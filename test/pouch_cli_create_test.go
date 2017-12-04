package main

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-check/check"
)

// PouchCreateSuite is the test suite fo help CLI.
type PouchCreateSuite struct {
}

func init() {
	check.Suite(&PouchCreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchCreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, IsLinux)
}

// TestPouchCreateName is to verify the correctness of creating contaier with specified name.
func (suite *PouchCreateSuite) TestPouchCreateName(c *check.C) {
	out, err := exec.Command("pouch", "create", "--name", "foo", "busybox:latest").Output()
	c.Assert(err, check.IsNil)

	if !strings.Contains(string(out), "foo") {
		c.Fatalf("unexpected output %s expected foo\n", string(out))
	}
}

// TestPouchCreateDuplicateContainerName is to verify duplicate container names.
func (suite *PouchCreateSuite) TestPouchCreateDuplicateContainerName(c *check.C) {
	containername := "duplicate"
	out, err := exec.Command("pouch", "create", "--name", containername, "busybox:latest").Output()
	c.Assert(err, check.IsNil)

	out, err = exec.Command("pouch", "create", "--name", containername, "busybox:latest").CombinedOutput()
	c.Assert(err, check.NotNil)

	if !strings.Contains(string(out), "already exist") {
		c.Fatalf("unexpected output %s expected already exist\n", string(out))
	}
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchCreateSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, IsLinux)

	// Pull test image
	cmd := exec.Command("pouch", "pull", testImage)
	cmd.Run()
}

// TearDownSuite does cleanup work in the end of each test suite.
func (suite *PouchCreateSuite) TearDownSuite(c *check.C) {
	// TODO: Remove test image
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchCreateSuite) TearDownTest(c *check.C) {
	// TODO add cleanup work
}

// TestCreateWorks tests "pouch create" work.
func (suite *PouchCreateSuite) TestCreateWorks(c *check.C) {

	// TODO: add wrong args.
	args := map[string]bool{
		"":             true,
		"-t":           true,
		"-v /tmp:/tmp": true,
	}

	for arg, ok := range args {
		cmd := exec.Command("pouch", "create", arg, testImage)
		out, _, err := runCmd(cmd)

		if ok {
			c.Assert(err, check.IsNil)
			match, _ := regexp.MatchString("container.*name.*", out)
			c.Assert(match, check.Equals, true)
		} else {
			c.Assert(err, check.NotNil)
		}
	}
	// TODO: clean the created container
}
