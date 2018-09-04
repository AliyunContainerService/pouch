package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchCommitSuite is the test suite for container commit.
type PouchCommitSuite struct{}

func init() {
	check.Suite(&PouchCommitSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchCommitSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	SkipIfFalse(c, environment.IsCRIUExist)
}

// TestCommitContainer tests commit a container successful.
func (suite *PouchCommitSuite) TestCommitContainer(c *check.C) {
	cname := "TestCommitContainer"
	image := "foo:bar"

	command.PouchRun("run", "--name", cname, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	ret := command.PouchRun("commit", cname, image)
	ret.Assert(c, icmd.Success)
	defer DelImageForceOk(c, image)

	imageID := strings.TrimSpace(ret.Stdout())
	command.PouchRun("image", "inspect", imageID).Assert(c, icmd.Success)
}

// TestCommitNewFile tests commit a container include new file.
func (suite *PouchCommitSuite) TestCommitNewFile(c *check.C) {
	cname := "TestCommitNewFile"
	image := "foo:newfile"

	command.PouchRun("run", "--name", cname, busyboxImage, "/bin/sh", "-c", "echo a > /foo").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	ret := command.PouchRun("commit", cname, image)
	ret.Assert(c, icmd.Success)
	imageID := strings.TrimSpace(ret.Stdout())
	defer DelImageForceOk(c, image)

	nname := "fromNewFile"
	ret = command.PouchRun("run", "--name", nname, imageID, "cat", "/foo")
	ret.Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "a\n")
	DelContainerForceMultyTime(c, nname)
}

// TestCommitHardLink tests commit a container include a hard link.
func (suite *PouchCommitSuite) TestCommitHardLink(c *check.C) {
	cname := "TestCommitHardLink"
	image := "foo:hardlink"

	ret := command.PouchRun("run", "-t", "--name", cname, "busybox", "sh", "-c", "touch file1 && ln file1 file2 && ls -di file1 file2")
	ret.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	output := ret.Stdout()
	splits := strings.SplitN(strings.TrimSpace(output), " ", 2)
	inode1 := splits[0]
	splits = strings.SplitAfterN(strings.TrimSpace(output), " ", 2)
	inode2 := strings.TrimSpace(splits[0])
	c.Assert(inode1, check.Equals, inode2)

	ret = command.PouchRun("commit", cname, image)
	defer DelImageForceOk(c, image)
	ret.Assert(c, icmd.Success)
	imageID := strings.TrimSpace(ret.Stdout())

	nname := "fromHardlink"
	ret = command.PouchRun("run", "-t", "--name", nname, imageID, "sh", "-c", "ls -di file1 file2")
	ret.Assert(c, icmd.Success)
	output = ret.Stdout()
	splits = strings.SplitN(strings.TrimSpace(output), " ", 2)
	inode1 = splits[0]
	splits = strings.SplitAfterN(strings.TrimSpace(output), " ", 2)
	inode2 = strings.TrimSpace(splits[0])
	c.Assert(inode1, check.Equals, inode2)
	DelContainerForceMultyTime(c, nname)
}

// TestCommitBindMount tests commit a container include new file.
func (suite *PouchCommitSuite) TestCommitBindMount(c *check.C) {
	cname := "TestCommitBindMount"
	image := "foo:mount"

	command.PouchRun("run", "--name", cname, "-v", "/dev/null:/tmp/h1", busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)

	ret := command.PouchRun("commit", cname, image)
	defer DelImageForceOk(c, image)
	ret.Assert(c, icmd.Success)
	imageID := strings.TrimSpace(ret.Stdout())

	nname := "fromMount"
	ret = command.PouchRun("run", "--name", nname, imageID, "ls", "/tmp/h1")
	ret.Assert(c, icmd.Success)
	DelContainerForceMultyTime(c, nname)
}
