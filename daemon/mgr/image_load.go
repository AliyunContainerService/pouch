package mgr

import (
	"context"
	"fmt"
	"io"

	"github.com/alibaba/pouch/pkg/multierror"
	"github.com/alibaba/pouch/pkg/reference"

	ociimage "github.com/containerd/containerd/images/oci"
	pkgerrors "github.com/pkg/errors"
)

// LoadImage loads images by the oci.v1 format tarstream.
func (mgr *ImageManager) LoadImage(ctx context.Context, imageName string, tarstream io.ReadCloser) error {
	defer tarstream.Close()

	namedRef, err := reference.Parse(imageName)
	if err != nil {
		return pkgerrors.Wrapf(err, "failed to parse image name %s", imageName)
	}

	// NOTE: in the image ocispec.v1, the org.opencontainers.image.ref.name
	// annotation represents a "tag" for image. For example, an image may
	// have a tag for different versions or builds of the software.
	// And containerd.importer will append ":" and annotation to the name
	// so that we don't allow imageName to contains any digest or tag
	// information, like foo/bar:latest:v1.2.
	if !reference.IsNamedOnly(namedRef) {
		return fmt.Errorf("the image name should not contains any digest or tag information")
	}

	importer := &ociimage.V1Importer{
		ImageName: imageName,
	}

	imgs, err := mgr.client.ImportImage(ctx, importer, tarstream)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to import image into containerd by tarstream")
	}

	// FIXME(fuwei): if the store fails to update reference cache, the daemon
	// may fail to load after restart.
	merrs := new(multierror.Multierrors)
	for _, img := range imgs {
		if err := mgr.StoreImageReference(ctx, img); err != nil {
			merrs.Append(fmt.Errorf("fail to store reference: %s: %v", img.Name(), err))
		}
	}

	if merrs.Size() != 0 {
		return fmt.Errorf("fails to load image: %v", merrs.Error())
	}
	return nil
}
