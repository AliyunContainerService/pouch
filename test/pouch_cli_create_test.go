package main

import (
	"os/exec"

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

// TestPouchCreateName is to verify the correctness of creating contaier with specified name.
func (suite *PouchCreateSuite) TestPouchCreateName(c *check.C) {
	var cmd PouchCmd

	args := []string{"create", "--name", "foo", testImage}
	cmd = PouchCmd{
		args:        args,
		result:      true,
		outContains: "foo",
	}
	RunCmd(c, &cmd)
}

// TestPouchCreateDuplicateContainerName is to verify duplicate container names.
func (suite *PouchCreateSuite) TestPouchCreateDuplicateContainerName(c *check.C) {
	containername := "duplicate"
	args := []string{"create", "--name", containername, testImage}

	var cmd PouchCmd

	cmd = PouchCmd{
		args:        args,
		result:      true,
		outContains: containername,
	}
	RunCmd(c, &cmd)

	cmd = PouchCmd{
		args:        args,
		result:      false,
		returnValue: 1,
		outContains: "already exist",
	}
	RunCmd(c, &cmd)
}

// TestCreateWorks tests "pouch create" work.
func (suite *PouchCreateSuite) TestCreateWorks(c *check.C) {

	args := []string{"create", testImage, "-t"}
	cmd := PouchCmd{
		args:   args,
		result: true,
	}
	RunCmd(c, &cmd)

	args = []string{"create", "-v", "/tmp:/tmp", testImage}
	cmd = PouchCmd{
		args:   args,
		result: true,
	}
	RunCmd(c, &cmd)

	// TODO: clean the created container
}
