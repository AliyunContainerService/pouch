package mgr

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/jsonstream"
	"github.com/alibaba/pouch/registry"

	"github.com/pkg/errors"
	"github.com/tchap/go-patricia/patricia"
)

// ImageMgr as an interface defines all operations against images.
type ImageMgr interface {
	// PullImage pulls images from specified registry.
	PullImage(ctx context.Context, image, tag string, out io.Writer) error

	// ListImages lists images stored by containerd.
	ListImages(ctx context.Context, filters string) ([]types.ImageInfo, error)

	// Search Images from specified registry.
	SearchImages(ctx context.Context, name string, registry string) ([]types.SearchResultItem, error)

	// GetImage gets image by image id or ref.
	GetImage(ctx context.Context, idOrRef string) (*types.ImageInfo, error)

	// RemoveImage deletes an image by reference.
	RemoveImage(ctx context.Context, image *types.ImageInfo, option *ImageRemoveOption) error
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

	cache *imageCache
}

// NewImageManager initializes a brand new image manager.
func NewImageManager(cfg *config.Config, client *ctrd.Client) (*ImageManager, error) {
	mgr := &ImageManager{
		DefaultRegistry: "docker.io",
		client:          client,
		cache:           newImageCache(),
	}

	if err := mgr.loadImages(); err != nil {
		return nil, err
	}
	return mgr, nil
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

	img, err := mgr.client.PullImage(ctx, image+":"+tag, stream)

	// wait goroutine to exit.
	<-wait

	if err != nil {
		return err
	}

	mgr.cache.put(&img)

	return nil
}

// ListImages lists images stored by containerd.
func (mgr *ImageManager) ListImages(ctx context.Context, filters string) ([]types.ImageInfo, error) {
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

// GetImage gets image by image id or ref.
func (mgr *ImageManager) GetImage(ctx context.Context, idOrRef string) (*types.ImageInfo, error) {
	return mgr.cache.get(idOrRef)
}

// RemoveImage deletes an image by reference.
func (mgr *ImageManager) RemoveImage(ctx context.Context, image *types.ImageInfo, option *ImageRemoveOption) error {
	if err := mgr.client.RemoveImage(ctx, image.Name); err != nil {
		return err
	}

	// image has been deleted, delete from image cache too.
	mgr.cache.remove(image)

	return nil
}

func (mgr *ImageManager) loadImages() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	images, err := mgr.client.ListImages(ctx)
	cancel()
	if err != nil {
		return err
	}

	for i := range images {
		mgr.cache.put(&images[i])
	}
	return nil
}

type imageCache struct {
	sync.Mutex
	refs map[string]*types.ImageInfo // store the mapping ref to image.
	ids  *patricia.Trie              // store the mapping id to image.
}

func newImageCache() *imageCache {
	return &imageCache{
		refs: make(map[string]*types.ImageInfo),
		ids:  patricia.NewTrie(),
	}
}

func (c *imageCache) put(image *types.ImageInfo) {
	c.Lock()
	defer c.Unlock()

	id := strings.TrimPrefix(image.ID, "sha256:")
	ref := image.Name

	c.refs[ref] = image
	c.ids.Insert(patricia.Prefix(id), image)
}

func (c *imageCache) get(idOrRef string) (*types.ImageInfo, error) {
	c.Lock()
	defer c.Unlock()

	image, ok := c.refs[idOrRef]
	if ok {
		return image, nil
	}

	var images []*types.ImageInfo

	fn := func(prefix patricia.Prefix, item patricia.Item) error {
		if image, ok := item.(*types.ImageInfo); ok {
			images = append(images, image)
		}
		return nil
	}

	if err := c.ids.VisitSubtree(patricia.Prefix(idOrRef), fn); err != nil {
		// the error does not occur.
		return nil, err
	}

	if len(images) > 1 {
		return nil, errors.Wrap(errtypes.ErrTooMany, "image: "+idOrRef)
	} else if len(images) == 0 {
		return nil, errors.Wrap(errtypes.ErrNotfound, "image: "+idOrRef)
	}

	return images[0], nil
}

func (c *imageCache) remove(image *types.ImageInfo) {
	c.Lock()
	defer c.Unlock()

	id := strings.TrimPrefix(image.ID, "sha256:")
	ref := image.Name

	delete(c.refs, ref)
	c.ids.Delete(patricia.Prefix(id))
}
