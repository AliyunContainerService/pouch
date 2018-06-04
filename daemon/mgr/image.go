package mgr

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/jsonstream"
	"github.com/alibaba/pouch/pkg/reference"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/containerd/containerd"
	ctrdmetaimages "github.com/containerd/containerd/images"
	digest "github.com/opencontainers/go-digest"
	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var deadlineLoadImagesAtBootup = time.Second * 10

// ImageMgr as an interface defines all operations against images.
type ImageMgr interface {
	// PullImage pulls images from specified registry.
	PullImage(ctx context.Context, ref string, authConfig *types.AuthConfig, out io.Writer) error

	// GetImage returns imageInfo by reference or id.
	GetImage(ctx context.Context, idOrRef string) (*types.ImageInfo, error)

	// ListImages lists images stored by containerd.
	ListImages(ctx context.Context, filter ...string) ([]types.ImageInfo, error)

	// Search Images from specified registry.
	SearchImages(ctx context.Context, name string, registry string) ([]types.SearchResultItem, error)

	// RemoveImage deletes an image by reference.
	RemoveImage(ctx context.Context, idOrRef string, force bool) error

	// AddTag creates target ref for source image.
	AddTag(ctx context.Context, sourceImage string, targetRef string) error

	// CheckReference returns imageID, actual reference and primary reference.
	CheckReference(ctx context.Context, idOrRef string) (digest.Digest, reference.Named, reference.Named, error)

	// LoadImage creates a set of images by tarstream.
	LoadImage(ctx context.Context, imageName string, tarstream io.ReadCloser) error
}

// ImageManager is an implementation of interface ImageMgr.
type ImageManager struct {
	// DefaultRegistry is the default registry of daemon.
	// When users do not specify image repo in image name,
	// daemon will automatically pull images with DefaultRegistry and DefaultNamespace.
	DefaultRegistry string

	// DefaultNamespace is the default namespace used in DefaultRegistry.
	DefaultNamespace string

	// client is a interface to the containerd client.
	// It is used to interact with containerd.
	client ctrd.APIClient

	// localStore is local cache of image reference information.
	localStore *imageStore
}

// NewImageManager initializes a brand new image manager.
func NewImageManager(cfg *config.Config, client ctrd.APIClient) (*ImageManager, error) {
	store, err := newImageStore()
	if err != nil {
		return nil, err
	}

	mgr := &ImageManager{
		DefaultRegistry:  cfg.DefaultRegistry,
		DefaultNamespace: cfg.DefaultRegistryNS,

		client:     client,
		localStore: store,
	}

	if err := mgr.updateLocalStore(); err != nil {
		return nil, err
	}
	return mgr, nil
}

// PullImage pulls images from specified registry.
func (mgr *ImageManager) PullImage(ctx context.Context, ref string, authConfig *types.AuthConfig, out io.Writer) error {
	newRef := addDefaultRegistryIfMissing(ref, mgr.DefaultRegistry, mgr.DefaultNamespace)
	namedRef, err := reference.Parse(newRef)
	if err != nil {
		return err
	}

	pctx, cancel := context.WithCancel(ctx)
	stream := jsonstream.New(out)
	wait := make(chan struct{})

	go func() {
		// wait stream to finish.
		defer cancel()
		stream.Wait()
		close(wait)
	}()

	namedRef = reference.TrimTagForDigest(reference.WithDefaultTagIfMissing(namedRef))
	img, err := mgr.client.PullImage(pctx, namedRef.String(), authConfig, stream)
	// wait goroutine to exit.
	<-wait
	if err != nil {
		return err
	}

	return mgr.storeImageReference(ctx, img)
}

// GetImage returns imageInfo by reference.
func (mgr *ImageManager) GetImage(ctx context.Context, idOrRef string) (*types.ImageInfo, error) {
	_, _, ref, err := mgr.CheckReference(ctx, idOrRef)
	if err != nil {
		return nil, err
	}

	img, err := mgr.client.GetImage(ctx, ref.String())
	if err != nil {
		return nil, err
	}

	imgInfo, err := mgr.containerdImageToImageInfo(ctx, img)
	if err != nil {
		return nil, err
	}
	return &imgInfo, nil
}

// ListImages lists images stored by containerd.
func (mgr *ImageManager) ListImages(ctx context.Context, filter ...string) ([]types.ImageInfo, error) {
	imgs, err := mgr.client.ListImages(ctx, filter...)
	if err != nil {
		return nil, err
	}

	imgInfosIndexByID := make(map[string]types.ImageInfo)
	for _, img := range imgs {
		imgCfg, err := img.Config(ctx)
		if err != nil {
			return nil, err
		}

		if _, ok := imgInfosIndexByID[imgCfg.Digest.String()]; ok {
			continue
		}

		imgInfo, err := mgr.containerdImageToImageInfo(ctx, img)
		if err != nil {
			return nil, err
		}
		imgInfosIndexByID[imgInfo.ID] = imgInfo

	}

	imgInfos := make([]types.ImageInfo, 0, len(imgInfosIndexByID))
	for _, v := range imgInfosIndexByID {
		imgInfos = append(imgInfos, v)
	}
	return imgInfos, nil
}

// SearchImages searches imaged from specified registry.
func (mgr *ImageManager) SearchImages(ctx context.Context, name string, registry string) ([]types.SearchResultItem, error) {
	// Directly send API calls towards specified registry
	return nil, errtypes.ErrNotImplemented
}

// RemoveImage deletes a reference.
//
// NOTE: if the reference is short ID or ID, should remove all the references.
func (mgr *ImageManager) RemoveImage(ctx context.Context, idOrRef string, force bool) error {
	id, namedRef, primaryRef, err := mgr.CheckReference(ctx, idOrRef)
	if err != nil {
		return err
	}

	// should remove all the references if the reference is ID (Named Only)
	// or Digest ID (Tagged Named)
	if reference.IsNamedOnly(namedRef) ||
		strings.HasPrefix(id.String(), namedRef.String()) {

		// NOTE: the user maybe use the following references to pull one image
		//
		//	busybox:1.25
		//	busybox@sha256:29f5d56d12684887bdfa50dcd29fc31eea4aaf4ad3bec43daf19026a7ce69912
		//
		// There are referencing to the same image. They have the same
		// locator though there are two primary references. For this
		// case, we should remove two primary references without force
		// option.
		//
		// However, if there is alias like localhost:5000/busybox:latest
		// as searchable reference, we cannot remove the image because
		// the searchable reference has different locator without force.
		// It's different reference from locator aspect.
		if !force && !uniqueLocatorReference(mgr.localStore.GetReferences(id)) {
			return fmt.Errorf("Unable to remove the image %q (must force) - image has serveral references", idOrRef)
		}

		for _, ref := range mgr.localStore.GetPrimaryReferences(id) {
			if err := mgr.client.RemoveImage(ctx, ref.String()); err != nil {
				return err
			}

			if err := mgr.localStore.RemoveReference(id, ref); err != nil {
				return err
			}
		}
		return nil
	}

	namedRef = reference.TrimTagForDigest(namedRef)
	// remove the image if the nameRef is primary reference
	if primaryRef.String() == namedRef.String() {
		if err := mgr.localStore.RemoveReference(id, primaryRef); err != nil {
			return err
		}

		return mgr.client.RemoveImage(ctx, primaryRef.String())
	}
	return mgr.localStore.RemoveReference(id, namedRef)
}

// AddTag adds the tag reference to the source image.
//
// NOTE(fuwei): AddTag hacks the containerd metadata boltdb, which we add the
// reference into the containerd metadata boltdb with the existing image content.
// It means that the "tag" is primary reference in the pouchd.
//
// For example,
//	pouch tag A B
//	pouch rmi A
//
// The B is still there.
func (mgr *ImageManager) AddTag(ctx context.Context, sourceImage string, targetTag string) error {
	targetTag = addDefaultRegistryIfMissing(targetTag, mgr.DefaultRegistry, mgr.DefaultNamespace)

	tagRef, err := parseTagReference(targetTag)
	if err != nil {
		return err
	}

	if err := mgr.validateTagReference(tagRef); err != nil {
		return err
	}

	ctrdImg, err := mgr.fetchContainerdImage(ctx, sourceImage)
	if err != nil {
		return err
	}

	// add the reference into memory
	cfg, err := ctrdImg.Config(ctx)
	if err != nil {
		return err
	}
	if err := mgr.addReferenceIntoStore(cfg.Digest, tagRef, ctrdImg.Target().Digest); err != nil {
		return err
	}

	// add the reference into containerd meta db
	_, err = mgr.client.CreateImageReference(ctx, ctrdmetaimages.Image{
		Name:   tagRef.String(),
		Target: ctrdImg.Target(),
	})
	return err
}

// CheckReference returns image ID and actual reference.
func (mgr *ImageManager) CheckReference(ctx context.Context, idOrRef string) (actualID digest.Digest, actualRef reference.Named, primaryRef reference.Named, err error) {
	var namedRef reference.Named

	namedRef, err = reference.Parse(idOrRef)
	if err != nil {
		return
	}

	// NOTE: we cannot add default registry for the idOrRef directly
	// because the idOrRef maybe short ID or ID. we should run search
	// without addDefaultRegistryIfMissing at first round.
	actualID, actualRef, err = mgr.localStore.Search(namedRef)
	if err != nil {
		if !errtypes.IsNotfound(err) {
			return
		}

		newIDOrRef := addDefaultRegistryIfMissing(idOrRef, mgr.DefaultRegistry, mgr.DefaultNamespace)
		if newIDOrRef == idOrRef {
			return
		}

		// ideally, the err should be nil
		namedRef, err = reference.Parse(newIDOrRef)
		if err != nil {
			return
		}

		actualID, actualRef, err = mgr.localStore.Search(namedRef)
		if err != nil {
			return
		}
	}

	// NOTE: if the actualRef is ID (Named Only) or Digest ID (Tagged Named)
	// the primaryRef is first one of primary reference
	if reference.IsNamedOnly(actualRef) ||
		strings.HasPrefix(actualID.String(), actualRef.String()) {

		refs := mgr.localStore.GetPrimaryReferences(actualID)
		if len(refs) == 0 {
			err = errtypes.ErrNotfound
			logrus.Errorf("one Image ID must have the primary references, but got nothing")
			return
		}

		primaryRef = refs[0]
	} else {
		primaryRef, err = mgr.localStore.GetPrimaryReference(actualRef)
		if err != nil {
			return
		}
	}
	return
}

// updateLocalStore updates the local store.
func (mgr *ImageManager) updateLocalStore() error {
	ctx, cancel := context.WithTimeout(context.Background(), deadlineLoadImagesAtBootup)
	defer cancel()

	imgs, err := mgr.client.ListImages(ctx)
	if err != nil {
		return err
	}

	for _, img := range imgs {
		if err := mgr.storeImageReference(ctx, img); err != nil {
			logrus.Warnf("failed to load the image reference into local store: %v", err)
		}
	}
	return nil
}

func (mgr *ImageManager) storeImageReference(ctx context.Context, img containerd.Image) error {
	imgCfg, err := img.Config(ctx)
	if err != nil {
		return err
	}

	namedRef, err := reference.Parse(img.Name())
	if err != nil {
		return err
	}

	return mgr.addReferenceIntoStore(imgCfg.Digest, namedRef, img.Target().Digest)
}

func (mgr *ImageManager) addReferenceIntoStore(id digest.Digest, ref reference.Named, dig digest.Digest) error {
	// add primary reference as searchable reference
	if err := mgr.localStore.AddReference(id, ref, ref); err != nil {
		return err
	}

	// add Name@Digest as searchable reference if the primary reference is Name:Tag
	if reference.IsNameTagged(ref) {
		// NOTE: The digest reference must be primary reference.
		// If the digest reference has been exist, it means that the
		// same image has been pulled successfully.
		digRef := reference.WithDigest(ref, dig)
		if _, _, err := mgr.localStore.Search(digRef); err != nil {
			if errtypes.IsNotfound(err) {
				return mgr.localStore.AddReference(id, ref, digRef)
			}
		}
	}
	return nil
}

func (mgr *ImageManager) containerdImageToImageInfo(ctx context.Context, img containerd.Image) (types.ImageInfo, error) {
	desc, err := img.Config(ctx)
	if err != nil {
		return types.ImageInfo{}, err
	}

	size, err := img.Size(ctx)
	if err != nil {
		return types.ImageInfo{}, err
	}

	ociImage, err := containerdImageToOciImage(ctx, img)
	if err != nil {
		return types.ImageInfo{}, err
	}

	var (
		repoTags    = make([]string, 0)
		repoDigests = make([]string, 0)
	)

	for _, ref := range mgr.localStore.GetReferences(desc.Digest) {
		switch ref.(type) {
		case reference.Tagged:
			repoTags = append(repoTags, ref.String())
		case reference.CanonicalDigested:
			repoDigests = append(repoDigests, ref.String())
		}
	}

	return types.ImageInfo{
		Architecture: ociImage.Architecture,
		Config:       getImageInfoConfigFromOciImage(ociImage),
		CreatedAt:    ociImage.Created.Format(utils.TimeLayout),
		ID:           desc.Digest.String(),
		Os:           ociImage.OS,
		RepoDigests:  repoDigests,
		RepoTags:     repoTags,
		RootFS: &types.ImageInfoRootFS{
			Type:   ociImage.RootFS.Type,
			Layers: digestSliceToStringSlice(ociImage.RootFS.DiffIDs),
		},
		Size: size,
	}, nil
}

func (mgr *ImageManager) fetchContainerdImage(ctx context.Context, idOrRef string) (containerd.Image, error) {
	_, _, ref, err := mgr.CheckReference(ctx, idOrRef)
	if err != nil {
		return nil, err
	}

	return mgr.client.GetImage(ctx, ref.String())
}

func (mgr *ImageManager) validateTagReference(ref reference.Named) error {
	if _, ok := ref.(reference.Digested); ok {
		return pkgerrors.Wrap(
			errtypes.ErrInvalidParam,
			fmt.Sprintf("target tag reference (%s) cannot contains any digest information", ref.String()),
		)
	}

	// NOTE: we don't allow to use tag to override the existing primary reference.
	pRef, err := mgr.localStore.GetPrimaryReference(ref)
	if err != nil {
		return nil
	}

	if pRef.String() == ref.String() {
		return pkgerrors.Wrap(
			errtypes.ErrInvalidParam,
			fmt.Sprintf("the tag reference (%s) has been used as reference", ref.String()),
		)
	}
	return nil
}

func parseTagReference(targetTag string) (reference.Named, error) {
	ref, err := reference.Parse(targetTag)
	if err != nil {
		return nil, pkgerrors.Wrap(errtypes.ErrInvalidParam, err.Error())
	}

	return reference.WithDefaultTagIfMissing(ref), nil
}
