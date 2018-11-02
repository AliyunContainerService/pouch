package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/pkg/errors"
)

// PouchImagesSuite is the test suite for images CLI.
type PouchImagesSuite struct{}

func init() {
	check.Suite(&PouchImagesSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchImagesSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)
	environment.PruneAllImages(apiClient)

	PullImage(c, busyboxImage)
	PullImage(c, helloworldImage)
}

// TestImagesWorks tests "pouch images" work.
func (suite *PouchImagesSuite) TestImagesWorks(c *check.C) {
	image, err := getImageInfo(apiClient, busyboxImage)
	c.Assert(err, check.IsNil)

	sortImageID := func(ids string) string {
		idSlice := strings.Fields(ids)
		sort.Strings(idSlice)
		return strings.Join(idSlice, "\n")
	}

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

		qIDs, quietIDs := sortImageID(resQ.Combined()), sortImageID(resQuiet.Combined())

		c.Assert(qIDs, check.Equals, quietIDs)
		err := util.PartialEqual(strings.TrimSpace(qIDs), utils.TruncateID(image.ID))
		c.Assert(err, check.IsNil)
	}

	// with --digest
	{
		res := command.PouchRun("images", "--digest").Assert(c, icmd.Success)
		items := imagesListToKV(res.Combined())[busyboxImage]
		c.Assert(items[2], check.Equals, strings.TrimPrefix(image.RepoDigests[0], environment.BusyboxRepo+"@"))
	}

	// with --no-trunc
	{
		res := command.PouchRun("images", "--no-trunc").Assert(c, icmd.Success)
		items := imagesListToKV(res.Combined())[busyboxImage]
		c.Assert(items[0], check.Equals, image.ID)
	}
}

//TestImageListFilter test the filter flag works right
func (suite *PouchImagesSuite) TestImageListFilter(c *check.C) {
	busyBoxImageInfo, err := getImageInfo(apiClient, busyboxImage)
	c.Assert(err, check.IsNil)

	//Test Reference filter
	referenceRes := command.PouchRun("images", "-f", "reference="+busyboxImage)
	items := imagesListToKV(referenceRes.Combined())
	c.Assert(len(items), check.Equals, 1)
	c.Assert(items[busyboxImage][1], check.Equals, busyBoxImageInfo.RepoTags[0])

	//Test before filter
	beforeRes1 := command.PouchRun("images", "-f", "before="+busyboxImage)
	items1 := imagesListToKV(beforeRes1.Combined())
	beforeRes2 := command.PouchRun("images", "-f", "before="+helloworldImage)
	items2 := imagesListToKV(beforeRes2.Combined())
	c.Assert(len(items1)+len(items2), check.Equals, 1)

	//Test since filter
	sinceRes1 := command.PouchRun("images", "-f", "since="+busyboxImage)
	items1 = imagesListToKV(sinceRes1.Combined())
	sinceRes2 := command.PouchRun("images", "-f", "since="+helloworldImage)
	items2 = imagesListToKV(sinceRes2.Combined())
	c.Assert(len(items1)+len(items2), check.Equals, 1)
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
	images, err := apiClient.ImageList(ctx, filters.NewArgs())
	if err != nil {
		return types.ImageInfo{}, errors.Wrap(err, "fail to list images")
	}

	for _, img := range images {
		if img.RepoTags[0] == name {
			return img, nil
		}
	}
	return types.ImageInfo{}, errors.Errorf("image %s not found", name)
}

// TestInspectImage is to verify the format flag of image inspect command.
func (suite *PouchImagesSuite) TestInspectImage(c *check.C) {
	output := command.PouchRun("image", "inspect", busyboxImage).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	// inspect image name
	output = command.PouchRun("image", "inspect", "-f", "{{.RepoTags}}", busyboxImage).Stdout()
	c.Assert(output, check.Equals, fmt.Sprintf("[%s]\n", busyboxImage))
}

// TestLoginAndLogout is to test login and logout command
func (suite *PouchImagesSuite) TestLoginAndLogout(c *check.C) {
	SkipIfFalse(c, environment.IsHubConnected)

	// test login a defined registry
	output := command.PouchRun("login", "-u", testHubUser, "-p", testHubPasswd, testHubAddress).Stdout()
	c.Assert(util.PartialEqual(output, "Login Succeeded"), check.IsNil)

	// test logout a defined registry
	output = command.PouchRun("logout", testHubAddress).Stdout()
	c.Assert(util.PartialEqual(output, "Remove login credential for registry"), check.IsNil)
}
