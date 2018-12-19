package ctrd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/diff"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/rootfs"
	"github.com/containerd/containerd/snapshots"
	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/identity"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	emptyGZLayer           = digest.Digest("sha256:4f4fb700ef54461cfa02571ae0db9a0dc1e0cdb5577484a6d75e68dc38e8acc1")
	containerdUncompressed = "containerd.io/uncompressed"
	manifestType           = images.MediaTypeDockerSchema2Manifest
	configType             = images.MediaTypeDockerSchema2Config
	layerType              = images.MediaTypeDockerSchema2LayerGzip
)

// CommitConfig defines options for committing a container image
type CommitConfig struct {

	// author
	Author string

	// comment
	Comment string

	// container config
	ContainerConfig *types.ContainerConfig

	// container ID
	ContainerID string

	// parent reference
	ParentReference string

	// reference
	Reference string

	// repository
	Repository string

	// containerd format image
	CImage containerd.Image

	// image-spec format image
	Image ocispec.Image
}

// Commit commits an image from a container.
func (c *Client) Commit(ctx context.Context, config *CommitConfig) (_ digest.Digest, err0 error) {
	// get a containerd client
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}
	client := wrapperCli.client

	// NOTE: make sure that gc scheduler doesn't remove content/snapshot during commmit
	ctx, done, err := client.WithLease(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create lease for commit")
	}
	defer done()

	var (
		sn     = client.SnapshotService(CurrentSnapshotterName(ctx))
		cs     = client.ContentStore()
		differ = client.DiffService()
	)

	// export new layer
	snapshot, err := c.GetSnapshot(ctx, config.ContainerID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get snapshot")
	}

	layer, diffID, err := exportLayer(ctx, snapshot.Name, sn, cs, differ)
	if err != nil {
		return "", errors.Wrap(err, "failed to export layer")
	}

	childImg := newChildImage(ctx, config, diffID)

	// create new snapshot for new layer
	rootfsID := identity.ChainID(childImg.RootFS.DiffIDs).String()
	if err = newSnapshot(ctx, rootfsID, config.Image, sn, differ, layer); err != nil {
		return "", err
	}

	defer func() {
		if err0 != nil {
			logrus.Warnf("remove snapshot %s cause commit image failed", rootfsID)
			client.SnapshotService(CurrentSnapshotterName(ctx)).Remove(ctx, rootfsID)
		}
	}()

	imgJSON, err := json.Marshal(childImg)
	if err != nil {
		return "", err
	}

	// new config descriptor
	configDesc := ocispec.Descriptor{
		MediaType: configType,
		Digest:    digest.FromBytes(imgJSON),
		Size:      int64(len(imgJSON)),
	}

	// get parent image layer descriptor
	pmfst, err := images.Manifest(ctx, cs, config.CImage.Target(), platforms.Default())
	if err != nil {
		return "", err
	}

	// new layer descriptor
	layers := append(pmfst.Layers, layer)
	labels := map[string]string{
		"containerd.io/gc.ref.content.0": configDesc.Digest.String(),
	}
	for i, l := range layers {
		labels[fmt.Sprintf("containerd.io/gc.ref.content.%d", i+1)] = l.Digest.String()
	}

	// new manifest descriptor
	mfst := ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		Config: configDesc,
		Layers: layers,
	}

	mfstJSON, err := json.MarshalIndent(mfst, "", "   ")
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal manifest")
	}

	mfstDigest := digest.FromBytes(mfstJSON)
	mfstDesc := ocispec.Descriptor{
		Digest: mfstDigest,
		Size:   int64(len(mfstJSON)),
	}

	desc := ocispec.Descriptor{
		MediaType: manifestType,
		Digest:    mfstDigest,
		Size:      int64(len(mfstJSON)),
	}

	// image create
	img := images.Image{
		Name:      config.Reference,
		Target:    desc,
		CreatedAt: time.Now(),
	}

	// register containerd image metadata.
	if _, err := client.ImageService().Update(ctx, img); err != nil {
		if !errdefs.IsNotFound(err) {
			return "", fmt.Errorf("failed to cover exist image %s", err)
		}

		if _, err := client.ImageService().Create(ctx, img); err != nil {
			return "", fmt.Errorf("failed to create new image %s", err)
		}
	}

	// write manifest content
	ref := mfstDigest.String()
	if err := content.WriteBlob(ctx, cs, ref, bytes.NewReader(mfstJSON), mfstDesc.Size, mfstDesc.Digest, content.WithLabels(labels)); err != nil {
		return "", errors.Wrapf(err, "error writing manifest blob %s", mfstDigest)
	}

	// write config content
	ref = configDesc.Digest.String()
	labelOpt := content.WithLabels(map[string]string{
		fmt.Sprintf("containerd.io/gc.ref.snapshot.%s", CurrentSnapshotterName(ctx)): rootfsID,
	})
	if err := content.WriteBlob(ctx, cs, ref, bytes.NewReader(imgJSON), configDesc.Size, configDesc.Digest, labelOpt); err != nil {
		return "", errors.Wrap(err, "error writing config blob")
	}

	// pouch record config descriptor digest as image id.
	return configDesc.Digest, nil
}

// export a new layer from a container
func exportLayer(ctx context.Context, name string, sn snapshots.Snapshotter, cs content.Store, differ diff.Differ) (ocispec.Descriptor, digest.Digest, error) {
	rwDesc, err := rootfs.Diff(ctx, name, sn, differ)
	if err != nil {
		return ocispec.Descriptor{}, digest.Digest(""), fmt.Errorf("failed to diff: %s", err)
	}

	info, err := cs.Info(ctx, rwDesc.Digest)
	if err != nil {
		return ocispec.Descriptor{}, digest.Digest(""), fmt.Errorf("failed to get exported layer info: %s", err)
	}

	diffIDStr, ok := info.Labels[containerdUncompressed]
	if !ok {
		return ocispec.Descriptor{}, digest.Digest(""), fmt.Errorf("invalid differ response with no diffID")
	}

	diffID, err := digest.Parse(diffIDStr)
	if err != nil {
		return ocispec.Descriptor{}, digest.Digest(""), err
	}

	layer := ocispec.Descriptor{
		MediaType: layerType,
		Digest:    rwDesc.Digest,
		Size:      info.Size,
	}
	return layer, diffID, nil
}

// create a new child image descriptor
func newChildImage(ctx context.Context, config *CommitConfig, diffID digest.Digest) ocispec.Image {
	createdTime := time.Now()
	emptyLayer := (diffID == emptyGZLayer)
	history := ocispec.History{
		Created:    &createdTime,
		CreatedBy:  strings.Join(config.ContainerConfig.Cmd, " "),
		Author:     config.Author,
		Comment:    config.Comment,
		EmptyLayer: emptyLayer,
	}

	// new child image
	pImg := config.Image
	return ocispec.Image{
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
		Created:      &createdTime,
		Author:       config.Author,
		Config:       newImageConfig(config.ContainerConfig),
		RootFS: ocispec.RootFS{
			Type:    "layers",
			DiffIDs: append(pImg.RootFS.DiffIDs, diffID),
		},
		History: append(pImg.History, history),
	}
}

// create a new snapshot for exported layer
func newSnapshot(ctx context.Context, name string, pImg ocispec.Image, sn snapshots.Snapshotter, differ diff.Differ, layer ocispec.Descriptor) error {
	var (
		key    = randomid.Generate()
		parent = identity.ChainID(pImg.RootFS.DiffIDs).String()
	)

	mount, err := sn.Prepare(ctx, key, parent,
		snapshots.WithLabels(map[string]string{
			snapshots.TypeLabelKey: snapshots.ImageType,
		}),
	)
	if err != nil {
		return err
	}

	// apply diff
	if _, err = differ.Apply(ctx, layer, mount); err != nil {
		return fmt.Errorf("failed to apply layer: %s", err)
	}

	if err = sn.Commit(ctx, name, key); err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return fmt.Errorf("failed to commit snapshot %s: %s", key, err)
		}

		// Destination already exists, cleanup key and return without error
		if err := sn.Remove(ctx, key); err != nil {
			return fmt.Errorf("failed to cleanup aborted apply %s: %s", key, err)
		}
	}
	return nil
}

// create a new image config descriptor
func newImageConfig(c *types.ContainerConfig) ocispec.ImageConfig {
	volumes := make(map[string]struct{})
	for i, v := range c.Volumes {
		if nv, ok := v.(struct{}); ok {
			volumes[i] = struct{}(nv)
		}
	}
	return ocispec.ImageConfig{
		User:       c.User,
		Env:        c.Env,
		Entrypoint: c.Entrypoint,
		Cmd:        c.Cmd,
		Volumes:    volumes,
		WorkingDir: c.WorkingDir,
		Labels:     c.Labels,
	}
}
