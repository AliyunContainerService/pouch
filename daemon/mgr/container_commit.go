package mgr

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/log"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// Commit commits an image from a container.
func (mgr *ContainerManager) Commit(ctx context.Context, name string, options *types.ContainerCommitOptions) (*types.ContainerCommitResp, error) {
	if options.Repository == "" {
		return nil, errors.Wrapf(errtypes.ErrInvalidParam, "not allow empty image reference")
	}
	if options.Tag == "" {
		options.Tag = "latest"
	}

	c, err := mgr.container(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find container(%s) to commit", name)
	}

	ctx = log.AddFields(ctx, map[string]interface{}{"ContainerID": c.ID})

	// Image keeps image digest name.
	_, _, pRef, err := mgr.ImageMgr.CheckReference(ctx, c.Image)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get container(%s) image", name)
	}

	// GetImage accepts reference like busybox:latest.
	img, err := mgr.Client.GetImage(ctx, pRef.String())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get image %s", c.Image)
	}
	ociImage, err := containerdImageToOciImage(ctx, img)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert containerd image to oci image")
	}
	// merge image's config into container
	if err := c.merge(func() (ocispec.ImageConfig, error) {
		return ociImage.Config, nil
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to merge config from image")
	}

	commitConfig := &ctrd.CommitConfig{
		Author:          options.Author,
		Comment:         options.Comment,
		ContainerID:     c.ID,
		Reference:       options.Repository + ":" + options.Tag,
		ParentReference: pRef.String(),
		ContainerConfig: c.Config,
		CImage:          img,
		Image:           ociImage,
	}

	// before image commit, call WithImageUnpack
	ctx = ctrd.WithImageUnpack(ctx)

	imageDigest, err := mgr.Client.Commit(ctx, commitConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to commit container (%s)", name)
	}

	// get new containerd image from containerd store.
	nImg, err := mgr.Client.GetImage(ctx, commitConfig.Reference)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get new created image %s from containerd", commitConfig.Reference)
	}

	// update image reference in pouch
	if err := mgr.ImageMgr.StoreImageReference(ctx, nImg); err != nil {
		// ignore error here, if failed to update image store only infect
		// pouch, but it does successful created a new image, restart
		// pouch can see the new image reference.
		log.With(ctx).Warnf("failed to update image store: %s", err)
	}
	mgr.LogContainerEvent(ctx, c, "commit")

	// return 12 bits image id as a result
	imageID := imageDigest.Hex()
	return &types.ContainerCommitResp{ID: string(imageID[:12])}, nil
}
