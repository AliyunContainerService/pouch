package main

import (
	"context"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/pkg/errors"
)

// PouchImagesSuite is the test suite fo help CLI.
type PouchImagesSuite struct{}

func init() {
	check.Suite(&PouchImagesSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchImagesSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TestImagesWorks tests "pouch image" work.
func (suite *PouchImagesSuite) TestImagesWorks(c *check.C) {
	image, err := getImageInfo(apiClient, busyboxImage)
	c.Assert(err, check.IsNil)

	// without flag
	{
		res := command.PouchRun("images").Assert(c, icmd.Success)
		items := imagesListToKV(res.Combined())[busyboxImage]

		c.Assert(items[0], check.Equals, utils.TruncateID(image.ID))
	}

	// with -q and --quiet
	{
		resQ := command.PouchRun("images", "-q").Assert(c, icmd.Success)
		resQuiet := command.PouchRun("images", "--quiet").Assert(c, icmd.Success)

		c.Assert(resQ.Combined(), check.Equals, resQuiet.Combined())
		c.Assert(strings.TrimSpace(resQ.Combined()), check.Equals, utils.TruncateID(image.ID))
	}

	// with --digest
	{
		res := command.PouchRun("images", "--digest").Assert(c, icmd.Success)
		items := imagesListToKV(res.Combined())[busyboxImage]
		c.Assert(items[2], check.Equals, image.Digest)
	}
}

// imagesListToKV parse "pouch images" into key-value mapping.
func imagesListToKV(list string) map[string][]string {
	// skip header
	lines := strings.Split(list, "\n")[1:]

	res := make(map[string][]string)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		items := strings.Fields(line)
		res[items[1]] = items
	}
	return res
}

// getImageInfo is used to retrieve the information about image.
func getImageInfo(apiClient client.ImageAPIClient, name string) (types.ImageInfo, error) {
	ctx := context.Background()
	images, err := apiClient.ImageList(ctx)
	if err != nil {
		return types.ImageInfo{}, errors.Wrap(err, "fail to list images")
	}

	for _, img := range images {
		if img.Name == name {
			return img, nil
		}
	}
	return types.ImageInfo{}, errors.Errorf("image %s not found", name)
}
