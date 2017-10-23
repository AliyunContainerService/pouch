package ctrd

import (
	"context"

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
func (c *Client) ListImages(ctx context.Context) ([]string, error) {
	images, err := c.client.ListImages(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list images")
	}
	list := make([]string, 0, len(images))
	for _, i := range images {
		list = append(list, i.Name())
	}
	return list, nil
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
