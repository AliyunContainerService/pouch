package mgr

import (
	"context"
	"io"

	ociimage "github.com/containerd/containerd/images/oci"
)

// SaveImage saves image to the oci.v1 format tarstream.
func (mgr *ImageManager) SaveImage(ctx context.Context, idOrRef string) (io.ReadCloser, error) {
	_, _, ref, err := mgr.CheckReference(ctx, idOrRef)
	if err != nil {
		return nil, err
	}

	exportedStream, err := mgr.client.SaveImage(ctx, &ociimage.V1Exporter{}, ref.String())
	if err != nil {
		return nil, err
	}
	mgr.LogImageEvent(ctx, idOrRef, ref.String(), "save")

	return exportedStream, nil
}
