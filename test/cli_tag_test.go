package main

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchTagSuite is the test suite for pouch tag.
type PouchTagSuite struct{}

func init() {
	check.Suite(&PouchTagSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchTagSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestImageTagOKWithSourceImageName tests OK.
func (suite *PouchTagSuite) TestImageTagOKWithSourceImageName(c *check.C) {
	repo, tag := "localhost:5000/testimagetagok/pouch", "source.name"
	tagRef := fmt.Sprintf("%s:%s", repo, tag)
	command.PouchRun("tag", busyboxImage, tagRef).Assert(c, icmd.Success)
	defer forceDeleteImage(c, tagRef)

	command.PouchRun("image", "inspect", tagRef).Assert(c, icmd.Success)
}

// TestImageTagOKWithSourceImageID tests OK.
func (suite *PouchTagSuite) TestImageTagOKWithSourceImageID(c *check.C) {
	repo, tag := "localhost:5000/testimagetagok/pouch", "source.id"
	tagRef := fmt.Sprintf("%s:%s", repo, tag)

	command.PouchRun("tag", environment.BusyboxID, tagRef).Assert(c, icmd.Success)
	defer forceDeleteImage(c, tagRef)

	command.PouchRun("image", "inspect", tagRef).Assert(c, icmd.Success)
}

// TestImageTagOKTargetWithoutTag tests fail.
func (suite *PouchTagSuite) TestImageTagOKTargetWithoutTag(c *check.C) {
	repo, tag := "localhost:5000/testimagetagok/pouch", "latest"
	tagRef := fmt.Sprintf("%s:%s", repo, tag)

	command.PouchRun("tag", busyboxImage, repo).Assert(c, icmd.Success)
	defer forceDeleteImage(c, tagRef)

	command.PouchRun("image", "inspect", tagRef).Assert(c, icmd.Success)
}

// TestImageTagFailToUseDigest tests fail.
func (suite *PouchTagSuite) TestImageTagFailToUseDigest(c *check.C) {
	repo, tag := "localhost:5000/testimagetagfail/pouch", "1.25"
	dig := "sha256:1ac48589692a53a9b8c2d1ceaa6b402665aa7fe667ba51ccc03002300856d8c7"

	tagRef := fmt.Sprintf("%s:%s@%s", repo, tag, dig)
	got := command.PouchRun("tag", busyboxImage, tagRef).Stderr()
	c.Assert(got, check.NotNil)

	expectedErr := "refusing to create a tag with a digest reference"
	if !strings.Contains(got, expectedErr) {
		c.Errorf("expected to contains %s, but got %v", expectedErr, got)
	}
}

// TestImageTagFailToUseSha256AsName tests fail.
func (suite *PouchTagSuite) TestImageTagFailToUseSha256AsName(c *check.C) {
	repo, tag := "localhost:5000/testimagetagfail/sha256", "1.25"

	tagRef := fmt.Sprintf("%s:%s", repo, tag)
	got := command.PouchRun("tag", busyboxImage, tagRef).Stderr()
	c.Assert(got, check.NotNil)

	expectedErr := "refusing to create an reference using digest algorithm as name"
	if !strings.Contains(got, expectedErr) {
		c.Errorf("expected to contains %s, but got %v", expectedErr, got)
	}
}
