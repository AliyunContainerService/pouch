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
	"github.com/alibaba/pouch/pkg/reference"
	"github.com/alibaba/pouch/registry"

	"github.com/pkg/errors"
	"github.com/tchap/go-patricia/patricia"
)

// ImageMgr as an interface defines all operations against images.
type ImageMgr interface {
	// PullImage pulls images from specified registry.
	PullImage(ctx context.Context, imageRef string, authConfig *types.AuthConfig, out io.Writer) error

	// ListImages lists images stored by containerd.
	ListImages(ctx context.Context, filters string) ([]types.ImageInfo, error)

	// Search Images from specified registry.
	SearchImages(ctx context.Context, name string, registry string) ([]types.SearchResultItem, error)

	// GetImage gets image by image id or ref.
	GetImage(ctx context.Context, idOrRef string) (*types.ImageInfo, error)

	// RemoveImage deletes an image by reference.
	RemoveImage(ctx context.Context, image *types.ImageInfo, name string, option *ImageRemoveOption) error
}

// ImageManager is an implementation of interface ImageMgr.
// It is a stateless manager, and it will never store image details.
// When image details needed from users, ImageManager interacts with
// containerd to get details.
type ImageManager struct {
	// DefaultRegistry is the default registry of daemon.
	// When users do not specify image repo in image name,
	// daemon will automatically pull images with DefaultRegistry and DefaultNamespace.
	DefaultRegistry string
	// DefaultNamespace is the default namespace used in DefaultRegistry.
	DefaultNamespace string

	// client is a interface to the containerd client.
	// It is used to interact with containerd.
	client   ctrd.APIClient
	registry *registry.Client

	cache *imageCache
}

// NewImageManager initializes a brand new image manager.
func NewImageManager(cfg *config.Config, client ctrd.APIClient) (*ImageManager, error) {
	mgr := &ImageManager{
		DefaultRegistry:  cfg.DefaultRegistry,
		DefaultNamespace: cfg.DefaultRegistryNS,
		client:           client,
		cache:            newImageCache(),
	}

	if err := mgr.loadImages(); err != nil {
		return nil, err
	}
	return mgr, nil
}

// PullImage pulls images from specified registry.
func (mgr *ImageManager) PullImage(pctx context.Context, imageRef string, authConfig *types.AuthConfig, out io.Writer) error {
	ctx, cancel := context.WithCancel(pctx)

	stream := jsonstream.New(out)
	wait := make(chan struct{})

	go func() {
		// wait stream to finish.
		stream.Wait()
		cancel()
		close(wait)
	}()

	imageRef = mgr.addRegistry(imageRef)
	img, err := mgr.client.PullImage(ctx, imageRef, authConfig, stream)

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
	return nil, errtypes.ErrNotImplemented
}

// GetImage gets image by image id or ref.
func (mgr *ImageManager) GetImage(ctx context.Context, idOrRef string) (*types.ImageInfo, error) {
	idOrRef = mgr.addRegistry(idOrRef)
	return mgr.cache.get(idOrRef)
}

// RemoveImage deletes an image by reference.
func (mgr *ImageManager) RemoveImage(ctx context.Context, image *types.ImageInfo, name string, option *ImageRemoveOption) error {
	// name is image ID
	if strings.HasPrefix(strings.TrimPrefix(image.ID, "sha256:"), name) {
		refs := mgr.cache.getIDToRefs(image.ID)
		mgr.cache.remove(image)
		return mgr.client.RemoveImage(ctx, refs[0])
	}

	// name is tagged or digest
	name = mgr.addRegistry(name)
	refNamed, err := reference.ParseNamedReference(name)
	if err != nil {
		return err
	}

	var ref string
	if refDigest, ok := refNamed.(reference.Digested); ok {
		ref = refDigest.String()
	}
	if refTagged, ok := refNamed.(reference.Tagged); ok {
		ref = refTagged.String()
	}

	if err := mgr.client.RemoveImage(ctx, ref); err != nil {
		return err
	}
	mgr.cache.untagged(refNamed)
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
	ids         *patricia.Trie             // store the mapping id to image.
	idToTags    map[string]map[string]bool // store repoTags by id, the repoTag is ref if bool is true.
	idToDigests map[string]map[string]bool // store repoDigests by id, the repoDigest is ref if bool is true.
	refToID     map[string]string          // store refs by id.
	tagToDigest map[string]string          // store the mapping repoTag to repoDigest.
	digestToTag map[string]string          // store the mapping repoDigest to repoTag,
}

func newImageCache() *imageCache {
	return &imageCache{
		ids:         patricia.NewTrie(),
		idToTags:    make(map[string]map[string]bool),
		idToDigests: make(map[string]map[string]bool),
		refToID:     make(map[string]string),
		tagToDigest: make(map[string]string),
		digestToTag: make(map[string]string),
	}
}

func (c *imageCache) put(image *types.ImageInfo) {
	c.Lock()
	defer c.Unlock()

	id := strings.TrimPrefix(image.ID, "sha256:")

	repoDigests := image.RepoDigests
	repoTags := image.RepoTags

	// NOTE: actually we simplify something, we assume that
	// tag and digest are one-to-one mapping and we can only
	// atmost one tag and one digest at a time.
	if len(repoTags) > 0 {
		// Pull with TagRef.
		if c.idToTags[id] == nil {
			c.idToTags[id] = make(map[string]bool)
		}
		c.idToTags[id][repoTags[0]] = true
		c.tagToDigest[repoTags[0]] = repoDigests[0]
		c.digestToTag[repoDigests[0]] = repoTags[0]
		c.refToID[repoTags[0]] = id
	}

	if c.idToDigests[id] == nil {
		c.idToDigests[id] = make(map[string]bool)
	}
	c.idToDigests[id][repoDigests[0]] = true
	c.refToID[repoDigests[0]] = id

	if item := c.ids.Get(patricia.Prefix(id)); item != nil {
		repoTags := []string{}
		repoDigests := []string{}
		for tag := range c.idToTags[id] {
			repoTags = append(repoTags, tag)
		}
		for digest := range c.idToDigests[id] {
			repoDigests = append(repoDigests, digest)
		}

		// Reset the image's RepoTags and RepoDigests.
		if img, ok := item.(*types.ImageInfo); ok {
			img.RepoTags = repoTags
			img.RepoDigests = repoDigests
			c.ids.Set(patricia.Prefix(id), img)
		}
	} else {
		c.ids.Insert(patricia.Prefix(id), image)
	}
}

func (c *imageCache) get(idOrRef string) (*types.ImageInfo, error) {
	c.Lock()
	defer c.Unlock()

	// use reference to parse idOrRef and add default tag if missing
	refNamed, err := reference.ParseNamedReference(idOrRef)
	if err != nil {
		return nil, err
	}

	var id string
	if refDigest, ok := refNamed.(reference.Digested); ok {
		id = c.refToID[refDigest.String()]
		if id == "" {
			return nil, errors.Wrap(errtypes.ErrNotfound, "image digest: "+refDigest.String())
		}
	} else {
		refTagged := reference.WithDefaultTagIfMissing(refNamed).(reference.Tagged)
		if c.refToID[refTagged.String()] == "" {
			// Maybe idOrRef is image ID.
			id = idOrRef
		} else {
			id = c.refToID[refTagged.String()]
		}
	}

	// use trie to fetch image
	var images []*types.ImageInfo

	fn := func(prefix patricia.Prefix, item patricia.Item) error {
		if image, ok := item.(*types.ImageInfo); ok {
			images = append(images, image)
		}
		return nil
	}

	if err := c.ids.VisitSubtree(patricia.Prefix(id), fn); err != nil {
		// the error does not occur.
		return nil, err
	}

	if len(images) > 1 {
		return nil, errors.Wrap(errtypes.ErrTooMany, "image: "+id)
	} else if len(images) == 0 {
		return nil, errors.Wrap(errtypes.ErrNotfound, "image: "+id)
	}

	return images[0], nil
}

func (c *imageCache) remove(image *types.ImageInfo) {
	c.Lock()
	defer c.Unlock()

	id := strings.TrimPrefix(image.ID, "sha256:")

	for _, v := range image.RepoTags {
		delete(c.refToID, v)
		delete(c.tagToDigest, v)
	}
	for _, v := range image.RepoDigests {
		delete(c.refToID, v)
		delete(c.digestToTag, v)
	}
	delete(c.idToTags, id)
	delete(c.idToDigests, id)

	c.ids.Delete(patricia.Prefix(id))
}

// untagged is used to remove the deleted repoTag or repoDigest from image.
func (c *imageCache) untagged(refNamed reference.Named) {
	c.Lock()
	defer c.Unlock()

	var ref, tag, digest string
	if refDigest, ok := refNamed.(reference.Digested); ok {
		ref = refDigest.String()
		digest = ref
		tag = c.digestToTag[digest]
	} else if refTagged, ok := reference.WithDefaultTagIfMissing(refNamed).(reference.Tagged); ok {
		ref = refTagged.String()
		tag = ref
		digest = c.tagToDigest[tag]
	}

	id := c.refToID[ref]
	delete(c.idToTags[id], tag)
	delete(c.idToDigests[id], digest)
	delete(c.refToID, tag)
	delete(c.refToID, digest)
	delete(c.tagToDigest, tag)
	delete(c.digestToTag, digest)

	if len(c.idToTags[id]) == 0 && len(c.idToDigests[id]) == 0 {
		c.ids.Delete(patricia.Prefix(id))
		return
	}

	// Delete the corresponding tag and digest from idToTags and idToDigests
	if item := c.ids.Get(patricia.Prefix(id)); item != nil {
		repoTags := []string{}
		repoDigests := []string{}
		for t := range c.idToTags[id] {
			if t == tag {
				continue
			}
			repoTags = append(repoTags, t)
		}
		for d := range c.idToDigests[id] {
			if d == digest {
				continue
			}
			repoDigests = append(repoDigests, d)
		}

		if img, ok := item.(*types.ImageInfo); ok {
			img.RepoTags = repoTags
			img.RepoDigests = repoDigests
			c.ids.Set(patricia.Prefix(id), img)
		}
	}
}

// getIDToRefs returns refs stored by ID index.
func (c *imageCache) getIDToRefs(id string) []string {
	c.Lock()
	defer c.Unlock()

	refs := []string{}
	id = strings.TrimPrefix(id, "sha256:")
	for tag, v := range c.idToTags[id] {
		if v {
			refs = append(refs, tag)
		}
	}
	for digest, v := range c.idToDigests[id] {
		if v {
			refs = append(refs, digest)
		}
	}

	return refs
}
