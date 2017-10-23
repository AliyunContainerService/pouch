package ctrd

import (
	"context"
	"strconv"

	"github.com/alibaba/pouch/apis/types"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// RemoveImage removes an image.
func (c *Client) RemoveImage(ctx context.Context, ref string) error {
	err := c.client.ImageService().Delete(ctx, ref)
	if err != nil {
		return errors.Wrap(err, "failed to remove image")
	}
	return nil
}

// ListImages lists all images.
func (c *Client) ListImages(ctx context.Context, filter ...string) ([]types.Image, error) {
	imageList, err := c.client.ListImages(ctx, filter...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list images")
	}

	images := make([]types.Image, 0, 32)
	digestPrefix := "sha256:"
	for _, image := range imageList {
		descriptor := image.Target()
		digest := []byte(descriptor.Digest)
		size := descriptor.Size

		images = append(images, types.Image{
			Name:   image.Name(),
			ID:     string(digest[len(digestPrefix) : len(digestPrefix)+12]),
			Digest: string(digest),
			Size:   formatSize(size),
		})
	}
	return images, nil
}

// PullImage downloads an image from the remote repository.
func (c *Client) PullImage(ctx context.Context, ref string, handle func(context.Context, ocispec.Descriptor) ([]ocispec.Descriptor, error)) error {
	options := []containerd.RemoteOpts{
		containerd.WithPullUnpack,
	}
	if handle != nil {
		handleWrap := func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
			if desc.MediaType != images.MediaTypeDockerSchema1Manifest {
				return handle(ctx, desc)
			}
			return nil, nil
		}
		options = append(options, containerd.WithImageHandler(images.HandlerFunc(handleWrap)))
	}

	img, err := c.pullImage(ctx, ref, options)
	if err != nil {
		return err
	}

	logrus.Infof("success to pull image: %s", img.Name())
	return nil
}

func (c *Client) pullImage(ctx context.Context, ref string, options []containerd.RemoteOpts) (containerd.Image, error) {
	img, err := c.client.Pull(ctx, ref, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull image")
	}

	return img, nil
}

func formatSize(size int64) string {
	return strconv.FormatInt(size, 10)
}
