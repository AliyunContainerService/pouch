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

	//
	CImage containerd.Image

	//
	Image ocispec.Image
}

// Commit commits a image from a container.
func (c *Client) Commit(ctx context.Context, config *CommitConfig) (imageID string, err0 error) {
	// get a containerd client
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}
	client := wrapperCli.client

	var (
		sn     = client.SnapshotService(defaultSnapshotterName)
		cs     = client.ContentStore()
		differ = client.DiffService()
	)

	// export new layer
	snapshot, err := c.GetSnapshot(ctx, config.ContainerID)
	if err != nil {
		return "", fmt.Errorf("failed to get snapshot: %s", err)
	}
	layer, diffIDStr, err := exportLayer(ctx, snapshot.Name, sn, cs, differ)
	if err != nil {
		return "", err
	}

	// create child image
	diffIDDigest, err := digest.Parse(diffIDStr)
	if err != nil {
		return "", err
	}

	childImg, err := newChildImage(ctx, config, diffIDDigest)
	if err != nil {
		return "", err
	}

	// create new snapshot for new layer
	snapshotKey := identity.ChainID(childImg.RootFS.DiffIDs).String()
	if err = newSnapshot(ctx, config.Image, sn, differ, layer, snapshotKey, diffIDStr); err != nil {
		return "", err
	}
	defer func() {
		if err0 != nil {
			logrus.Warnf("remove snapshot %s cause commit image failed", snapshotKey)
			client.SnapshotService(defaultSnapshotterName).Remove(ctx, snapshotKey)
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
	mfst := struct {
		// MediaType is reserved in the OCI spec but
		// excluded from go types.
		MediaType string `json:"mediaType,omitempty"`
		ocispec.Manifest
	}{
		MediaType: manifestType,
		Manifest: ocispec.Manifest{
			Versioned: specs.Versioned{
				SchemaVersion: 2,
			},
			Config: configDesc,
			Layers: layers,
		},
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

	// write manifest content
	if err := content.WriteBlob(ctx, cs, mfstDigest.String(), bytes.NewReader(mfstJSON), mfstDesc.Size, mfstDesc.Digest, content.WithLabels(labels)); err != nil {
		return "", errors.Wrapf(err, "error writing manifest blob %s", mfstDigest)
	}

	// write config content
	if err := content.WriteBlob(ctx, cs, configDesc.Digest.String(), bytes.NewReader(imgJSON), configDesc.Size, configDesc.Digest); err != nil {
		return "", errors.Wrap(err, "error writing config blob")
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

	// pouch record config descriptor digest as image id.
	return configDesc.Digest.String(), nil
}

// export a new layer from a container
func exportLayer(ctx context.Context, name string, sn snapshots.Snapshotter, cs content.Store, differ diff.Differ) (ocispec.Descriptor, string, error) {
	// export new layer
	rwDesc, err := rootfs.Diff(ctx, name, sn, differ)
	if err != nil {
		return ocispec.Descriptor{}, "", fmt.Errorf("failed to diff: %s", err)
	}

	info, err := cs.Info(ctx, rwDesc.Digest)
	if err != nil {
		return ocispec.Descriptor{}, "", err
	}
	diffIDStr, ok := info.Labels[containerdUncompressed]
	if !ok {
		return ocispec.Descriptor{}, "", fmt.Errorf("invalid differ response with no diffID")
	}

	layer := ocispec.Descriptor{
		MediaType: layerType,
		Digest:    rwDesc.Digest,
		Size:      info.Size,
	}

	return layer, diffIDStr, nil
}

// create a new child image descriptor
func newChildImage(ctx context.Context, config *CommitConfig, diffIDDigest digest.Digest) (ocispec.Image, error) {
	createdTime := time.Now()
	emptyLayer := (diffIDDigest == emptyGZLayer)
	history := ocispec.History{
		Created:    &createdTime,
		CreatedBy:  strings.Join(config.ContainerConfig.Cmd, " "),
		Author:     config.Author,
		Comment:    config.Comment,
		EmptyLayer: emptyLayer,
	}

	// new child image
	pImg := config.Image
	image := ocispec.Image{
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
		Created:      &createdTime,
		Author:       config.Author,
		Config:       newImageConfig(config.ContainerConfig),
		RootFS: ocispec.RootFS{
			Type:    "layers",
			DiffIDs: append(pImg.RootFS.DiffIDs, diffIDDigest),
		},
		History: append(pImg.History, history),
	}

	return image, nil
}

// create a new snapshot for exported layer
func newSnapshot(ctx context.Context, pImg ocispec.Image, sn snapshots.Snapshotter, differ diff.Differ, layer ocispec.Descriptor, name, diffIDStr string) error {
	diffIDs := pImg.RootFS.DiffIDs
	parent := identity.ChainID(diffIDs).String()

	key := randomid.Generate()
	mount, err := sn.Prepare(ctx, key, parent)
	if err != nil {
		return err
	}

	// apply diff
	_, err = differ.Apply(ctx, layer, mount)

	withLabels := func(info *snapshots.Info) error {
		info.Labels = map[string]string{
			containerdUncompressed:  diffIDStr,
			"containerd.io/gc.root": time.Now().UTC().Format(time.RFC3339Nano),
		}
		return nil
	}

	if err = sn.Commit(ctx, name, key, withLabels); err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return fmt.Errorf("failed to commit snapshot %s: %s", key, err)
		}

		// Destination already exists, cleanup key and return without error
		err = nil
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
