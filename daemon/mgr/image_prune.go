package mgr

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

// PruneImage delete unused images
func (mgr *ImageManager) PruneImage(ctx context.Context, imagesInfo []types.ImageInfo, filter filters.Args) (types.ImagePruneResp, error) {
	var result types.ImagePruneResp

	if len(imagesInfo) == 0 {
		return result, nil
	}

	for _, imgInfo := range imagesInfo {
		img, err := mgr.localStore.GetCtrdImageInfo(digest.Digest(imgInfo.ID))
		if err != nil {
			logrus.Warnf("failed to get containerd image from ImageInfo during list images: %v", imgInfo.ID, err)
			break
		}

		//  build return data structure
		deleted := fmt.Sprintf("%s\n", img.ID.String())
		for _, diff := range img.OCISpec.RootFS.DiffIDs {
			deleted += fmt.Sprintf("%s\n", diff.String())
		}

		untagged := ""
		for _, tags := range imgInfo.RepoTags {
			untagged += fmt.Sprintf("%s\n", tags)
		}

		result.ImagesDeleted = append(result.ImagesDeleted, &types.ImageDeleteRespItem{
			Deleted:  deleted,
			Untagged: untagged,
		})

		if imgInfo.Size > 0 {
			result.SpaceReclaimed += imgInfo.Size
		}

		// delete image
		mgr.RemoveImage(ctx, imgInfo.ID, true)
	}

	return result, nil
}
