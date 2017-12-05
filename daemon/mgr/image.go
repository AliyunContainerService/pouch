package mgr

import (
	"context"
	"io"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/jsonstream"
	"github.com/alibaba/pouch/registry"
)

// ImageMgr as an interface defines all operations against images.
type ImageMgr interface {
	// PullImage pulls images from specified registry.
	PullImage(ctx context.Context, image, tag string, out io.Writer) error

	// ListImages lists images stored by containerd.
	ListImages(ctx context.Context, filters string) ([]types.Image, error)

	// Search Images from specified registry.
	SearchImages(ctx context.Context, name string, registry string) ([]types.SearchResultItem, error)

	// RemoveImages remove a image by reference in local repo
	RemoveImages(ctx context.Context, name string, option *ImageRemoveOption) error
}

// ImageManager is an implementation of interface ImageMgr.
// It is a stateless manager, and it will never store image details.
// When image details needed from users, ImageManager interacts with
// containerd to get details.
type ImageManager struct {
	// DefaultRegistry is the default registry of daemon.
	// When users do not specify image repo in image name,
	// daemon will automatically pull images from DefaultRegistry.
	// TODO: make DefaultRegistry can be reloaded.
	DefaultRegistry string

	// client is a pointer to the containerd client.
	// It is used to interact with containerd.
	client   *ctrd.Client
	registry *registry.Client
}

// NewImageManager initializes a brand new image manager.
func NewImageManager(cfg *config.Config, client *ctrd.Client) (*ImageManager, error) {
	return &ImageManager{
		DefaultRegistry: "docker.io",
		client:          client,
	}, nil
}

// PullImage pulls images from specified registry.
func (mgr *ImageManager) PullImage(pctx context.Context, image, tag string, out io.Writer) error {
	ctx, cancel := context.WithCancel(pctx)

	stream := jsonstream.New(out)
	wait := make(chan struct{})

	go func() {
		// wait stream to finish.
		stream.Wait()
		cancel()
		close(wait)
	}()

	err := mgr.client.PullImage(ctx, image+":"+tag, stream)

	// wait goroutine to exit.
	<-wait

	return err
}

// ListImages lists images stored by containerd.
func (mgr *ImageManager) ListImages(ctx context.Context, filters string) ([]types.Image, error) {
	imageList, err := mgr.client.ListImages(ctx, filters)
	if err != nil {
		return nil, err
	}

	return imageList, nil
}

// SearchImages searches imaged from specified registry.
func (mgr *ImageManager) SearchImages(ctx context.Context, name string, registry string) ([]types.SearchResultItem, error) {
	// Directly send API calls towards specified registry
	if registry == "" {
		registry = mgr.DefaultRegistry
	}

	return nil, nil
}

// ImageRemoveOption wraps the image remove interface params.
type ImageRemoveOption struct {
	Force bool
}

// RemoveImage delete image by reference
func (mgr *ImageManager) RemoveImages(ctx context.Context, name string, option *ImageRemoveOption) error {
	err := mgr.client.RemoveImages(ctx, name)
	return err
}
