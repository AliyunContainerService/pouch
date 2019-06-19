package ctrd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/jsonstream"
	"github.com/alibaba/pouch/pkg/reference"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	ctrdmetaimages "github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/schema1"
	"github.com/containerd/containerd/snapshots"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/identity"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CreateImageReference creates the image in the meta data in the containerd.
func (c *Client) CreateImageReference(ctx context.Context, img ctrdmetaimages.Image) (ctrdmetaimages.Image, error) {
	image, err := c.createImageReference(ctx, img)
	if err != nil {
		return image, convertCtrdErr(err)
	}
	return image, nil
}

// createImageReference creates the image in the meta data in the containerd.
func (c *Client) createImageReference(ctx context.Context, img ctrdmetaimages.Image) (ctrdmetaimages.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return ctrdmetaimages.Image{}, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	return wrapperCli.client.ImageService().Create(ctx, img)
}

// GetImage returns the containerd's Image.
func (c *Client) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	img, err := c.getImage(ctx, ref)
	if err != nil {
		return img, convertCtrdErr(err)
	}
	return img, nil
}

// getImage returns the containerd's Image.
func (c *Client) getImage(ctx context.Context, ref string) (containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	return wrapperCli.client.GetImage(ctx, ref)
}

// ListImages lists all images.
func (c *Client) ListImages(ctx context.Context, filter ...string) ([]containerd.Image, error) {
	imgs, err := c.listImages(ctx, filter...)
	if err != nil {
		return imgs, convertCtrdErr(err)
	}
	return imgs, nil
}

// listImages lists all images.
func (c *Client) listImages(ctx context.Context, filter ...string) ([]containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	return wrapperCli.client.ListImages(ctx, filter...)
}

// RemoveImage deletes an image.
func (c *Client) RemoveImage(ctx context.Context, ref string) error {
	if err := c.removeImage(ctx, ref); err != nil {
		return convertCtrdErr(err)
	}
	return nil
}

// removeImage deletes an image.
func (c *Client) removeImage(ctx context.Context, ref string) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	if err := wrapperCli.client.ImageService().Delete(ctx, ref); err != nil {
		return errors.Wrap(err, "failed to remove image")
	}
	return nil
}

// SaveImage saves image to tarstream
func (c *Client) SaveImage(ctx context.Context, exporter ctrdmetaimages.Exporter, ref string) (io.ReadCloser, error) {
	r, err := c.saveImage(ctx, exporter, ref)
	if err != nil {
		return r, convertCtrdErr(err)
	}
	return r, nil
}

// saveImage saves image to tarstream
func (c *Client) saveImage(ctx context.Context, exporter ctrdmetaimages.Exporter, ref string) (io.ReadCloser, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	image, err := c.GetImage(ctx, ref)
	if err != nil {
		return nil, err
	}

	desc := image.Target()
	// add annotations in image description
	if desc.Annotations == nil {
		desc.Annotations = make(map[string]string)
	}
	if s, exist := desc.Annotations[ocispec.AnnotationRefName]; !exist || s == "" {
		namedRef, err := reference.Parse(ref)
		if err != nil {
			return nil, err
		}

		if reference.IsNameTagged(namedRef) {
			desc.Annotations[ocispec.AnnotationRefName] = namedRef.(reference.Tagged).Tag()
		}
	}

	return wrapperCli.client.Export(ctx, exporter, desc)
}

// ImportImage creates a set of images by tarstream.
//
// NOTE: One tar may have several manifests.
func (c *Client) ImportImage(ctx context.Context, reader io.Reader, opts ...containerd.ImportOpt) ([]containerd.Image, error) {
	imgs, err := c.importImage(ctx, reader, opts...)
	if err != nil {
		return imgs, convertCtrdErr(err)
	}
	return imgs, nil
}

// importImage creates a set of images by tarstream.
//
// NOTE: One tar may have several manifests.
func (c *Client) importImage(ctx context.Context, reader io.Reader, opts ...containerd.ImportOpt) ([]containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	// NOTE: The import will store the data into boltdb. But the unpack may
	// fail. It is not transaction.
	imgs, err := wrapperCli.client.Import(ctx, reader, opts...)
	if err != nil {
		return nil, err
	}

	var (
		res         = make([]containerd.Image, 0, len(imgs))
		snapshotter = CurrentSnapshotterName(ctx)
	)

	for _, img := range imgs {
		image := containerd.NewImage(wrapperCli.client, img)

		err = image.Unpack(ctx, snapshotter)
		if err != nil {
			return nil, err
		}

		res = append(res, image)
	}
	return res, nil
}

// PushImage pushes image to registry
func (c *Client) PushImage(ctx context.Context, ref string, authConfig *types.AuthConfig, out io.Writer) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	img, err := wrapperCli.client.GetImage(ctx, ref)
	if err != nil {
		return convertCtrdErr(err)
	}

	pushTracker := docker.NewInMemoryTracker()

	resolver, _, err := c.getResolver(ctx, authConfig, ref, []string{ref}, docker.ResolverOptions{
		Tracker: pushTracker,
	})
	if err != nil {
		return err
	}

	ongoing := jsonstream.NewPushJobs(pushTracker)
	handler := ctrdmetaimages.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		ongoing.Add(remotes.MakeRefKey(ctx, desc))
		return nil, nil
	})

	// fetch progress status, then send to client via out channel.
	stream := jsonstream.New(out, nil)
	pctx, cancelProgress := context.WithCancel(ctx)
	wait := make(chan struct{})
	go func() {
		jsonstream.PushProcess(pctx, ongoing, stream)
		close(wait)
	}()

	err = wrapperCli.client.Push(ctx, ref, img.Target(),
		containerd.WithResolver(resolver),
		containerd.WithImageHandler(handler))

	cancelProgress()
	<-wait

	defer func() {
		stream.Close()
		stream.Wait()
	}()

	if err != nil {
		stream.WriteObject(jsonstream.JSONMessage{
			Error: &jsonstream.JSONError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
			ErrorMessage: err.Error(),
		})

		return err
	}

	logrus.Infof("push image %s successfully", ref)

	return nil
}

// PullImage fetches image content from the remote repository, and then unpacks into snapshotter
func (c *Client) PullImage(ctx context.Context, name string, refs []string, authConfig *types.AuthConfig, stream *jsonstream.JSONStream) (containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	// NOTE: make sure that gc scheduler doesn't remove content/snapshot during pull
	ctx, done, err := wrapperCli.client.WithLease(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create lease for commit")
	}
	defer done(ctx)

	resolver, availableRef, err := c.getResolver(ctx, authConfig, name, refs, docker.ResolverOptions{})
	if err != nil {
		return nil, err
	}

	logrus.Infof("pulling image name %v reference %v", name, availableRef)

	ongoing := newJobs(name)

	options := []containerd.RemoteOpt{
		containerd.WithSchema1Conversion,
		containerd.WithResolver(resolver),
	}

	handle := func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		if desc.MediaType != ctrdmetaimages.MediaTypeDockerSchema1Manifest {
			ongoing.add(desc)
		}
		return nil, nil
	}
	options = append(options, containerd.WithImageHandler(ctrdmetaimages.HandlerFunc(handle)))

	// fetch progress status, then send to client via out channel.
	pctx, cancelProgress := context.WithCancel(ctx)
	wait := make(chan struct{})

	go func() {
		if err := c.fetchProgress(pctx, wrapperCli, ongoing, stream); err != nil {
			logrus.Errorf("failed to get pull's progress: %v", err)
		}
		close(wait)

		logrus.Infof("fetch progress exited, ref: %s.", availableRef)
	}()

	// start to pull image.
	img, err := c.fetchImage(ctx, wrapperCli, availableRef, options)

	// cancel fetch progress before handle error.
	cancelProgress()

	// wait fetch progress to finish.
	<-wait

	if err != nil {
		return nil, err
	}

	logrus.Infof("success to fetch image: %s", img.Name())

	// before image unpack, call WithImageUnpack
	ctx = WithImageUnpack(ctx)

	// unpack image
	if err = img.Unpack(ctx, CurrentSnapshotterName(ctx)); err != nil {
		return nil, err
	}

	return img, nil
}

func (c *Client) fetchImage(ctx context.Context, wrapperCli *WrapperClient, ref string, options []containerd.RemoteOpt) (containerd.Image, error) {
	ctrdClient := wrapperCli.client
	pullCtx := &containerd.RemoteContext{}
	for _, o := range options {
		if err := o(ctrdClient, pullCtx); err != nil {
			return nil, err
		}
	}
	// use the default platform
	pullCtx.PlatformMatcher = platforms.Default()

	ctx, done, err := ctrdClient.WithLease(ctx)
	if err != nil {
		return nil, err
	}
	defer done(ctx)

	store := ctrdClient.ContentStore()
	name, desc, err := pullCtx.Resolver.Resolve(ctx, ref)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve reference %q", ref)
	}

	fetcher, err := pullCtx.Resolver.Fetcher(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get fetcher for %q", name)
	}

	var (
		handler ctrdmetaimages.Handler

		isConvertible bool
		converterFunc func(context.Context, ocispec.Descriptor) (ocispec.Descriptor, error)
	)

	if desc.MediaType == ctrdmetaimages.MediaTypeDockerSchema1Manifest && pullCtx.ConvertSchema1 {
		schema1Converter := schema1.NewConverter(store, fetcher)

		handler = ctrdmetaimages.Handlers(append(pullCtx.BaseHandlers, schema1Converter)...)

		isConvertible = true

		converterFunc = func(ctx context.Context, _ ocispec.Descriptor) (ocispec.Descriptor, error) {
			return schema1Converter.Convert(ctx)
		}
	} else {
		// Get all the children for a descriptor
		childrenHandler := ChildrenHandler(store, ctrdClient.SnapshotService(CurrentSnapshotterName(ctx)))
		// Set any children labels for that content
		childrenHandler = SetChildrenLabels(store, childrenHandler)
		// Filter children by platforms
		childrenHandler = ctrdmetaimages.FilterPlatforms(childrenHandler, pullCtx.PlatformMatcher)
		// Sort and limit manifests if a finite number is needed
		childrenHandler = ctrdmetaimages.LimitManifests(childrenHandler, pullCtx.PlatformMatcher, 1)

		// set isConvertible to true if there is application/octet-stream media type
		convertibleHandler := ctrdmetaimages.HandlerFunc(
			func(_ context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
				if desc.MediaType == docker.LegacyConfigMediaType {
					isConvertible = true
				}

				return []ocispec.Descriptor{}, nil
			},
		)

		handler = ctrdmetaimages.Handlers(append(pullCtx.BaseHandlers,
			remotes.FetchHandler(store, fetcher),
			convertibleHandler,
			childrenHandler,
		)...)

		converterFunc = func(ctx context.Context, desc ocispec.Descriptor) (ocispec.Descriptor, error) {
			return docker.ConvertManifest(ctx, store, desc)
		}
	}

	if err := ctrdmetaimages.Dispatch(ctx, handler, desc); err != nil {
		return nil, err
	}

	if isConvertible {
		if desc, err = converterFunc(ctx, desc); err != nil {
			return nil, err
		}
	}

	img := ctrdmetaimages.Image{
		Name:   name,
		Target: desc,
		Labels: pullCtx.Labels,
	}

	is := ctrdClient.ImageService()
	for {
		if created, err := is.Create(ctx, img); err != nil {
			if !errdefs.IsAlreadyExists(err) {
				return nil, err
			}

			updated, err := is.Update(ctx, img)
			if err != nil {
				// if image was removed, try create again
				if errdefs.IsNotFound(err) {
					continue
				}
				return nil, err
			}

			img = updated
		} else {
			img = created
		}

		i := containerd.NewImageWithPlatform(ctrdClient, img, pullCtx.PlatformMatcher)
		return i, nil
	}
}

// ChildrenHandler returns the immediate children of content described by the descriptor.
func ChildrenHandler(provider content.Provider, sn snapshots.Snapshotter) ctrdmetaimages.HandlerFunc {
	return func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		var descs []ocispec.Descriptor
		switch desc.MediaType {
		case ctrdmetaimages.MediaTypeDockerSchema2Manifest, ocispec.MediaTypeImageManifest:
			p, err := content.ReadBlob(ctx, provider, desc)
			if err != nil {
				return nil, err
			}

			// TODO(stevvooe): We just assume oci manifest, for now. There may be
			// subtle differences from the docker version.
			var manifest ocispec.Manifest
			if err := json.Unmarshal(p, &manifest); err != nil {
				return nil, err
			}

			descs = append(descs, manifest.Config)

			// if a snapshoot is already ok, then not fetch the tar
			var (
				chain []digest.Digest
				i     = 0
			)
			for _, layer := range manifest.Layers {
				chainID := identity.ChainID(append(chain, layer.Digest)).String()
				if _, err := sn.Stat(ctx, chainID); err != nil {
					if !errdefs.IsNotFound(err) {
						return nil, errors.Wrapf(err, "failed to stat snapshot %s", chainID)
					}
					break
				}
				i++
				chain = append(chain, layer.Digest)
			}

			descs = append(descs, manifest.Layers[i:]...)
		case ctrdmetaimages.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
			p, err := content.ReadBlob(ctx, provider, desc)
			if err != nil {
				return nil, err
			}

			var index ocispec.Index
			if err := json.Unmarshal(p, &index); err != nil {
				return nil, err
			}

			descs = append(descs, index.Manifests...)
		case ctrdmetaimages.MediaTypeDockerSchema2Layer, ctrdmetaimages.MediaTypeDockerSchema2LayerGzip,
			ctrdmetaimages.MediaTypeDockerSchema2LayerForeign, ctrdmetaimages.MediaTypeDockerSchema2LayerForeignGzip,
			ctrdmetaimages.MediaTypeDockerSchema2Config, ocispec.MediaTypeImageConfig,
			ocispec.MediaTypeImageLayer, ocispec.MediaTypeImageLayerGzip,
			ocispec.MediaTypeImageLayerNonDistributable, ocispec.MediaTypeImageLayerNonDistributableGzip,
			ctrdmetaimages.MediaTypeContainerd1Checkpoint, ctrdmetaimages.MediaTypeContainerd1CheckpointConfig:
			// childless data types.
			return nil, nil
		default:
			logrus.Warnf("encountered unknown type %v; children may not be fetched", desc.MediaType)
		}

		return descs, nil
	}
}

// SetChildrenLabels is a handler wrapper which sets labels for the content on
// the children returned by the handler and passes through the children.
// Must follow a handler that returns the children to be labeled.
func SetChildrenLabels(manager content.Manager, f ctrdmetaimages.HandlerFunc) ctrdmetaimages.HandlerFunc {
	return func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		children, err := f(ctx, desc)
		if err != nil {
			return children, err
		}

		if len(children) > 0 {
			info := content.Info{
				Digest: desc.Digest,
				Labels: map[string]string{},
			}
			fields := []string{}
			for i, ch := range children {
				// only store config and manifest
				if ch.MediaType != ctrdmetaimages.MediaTypeDockerSchema2Config &&
					ch.MediaType != ocispec.MediaTypeImageConfig &&
					ch.MediaType != ctrdmetaimages.MediaTypeDockerSchema2Manifest &&
					ch.MediaType != ocispec.MediaTypeImageManifest {
					continue
				}
				info.Labels[fmt.Sprintf("containerd.io/gc.ref.content.%d", i)] = ch.Digest.String()
				fields = append(fields, fmt.Sprintf("labels.containerd.io/gc.ref.content.%d", i))
			}

			_, err := manager.Update(ctx, info, fields...)
			if err != nil {
				return nil, err
			}
		}

		return children, err
	}
}

// FIXME(fuwei): put the fetchProgress into jsonstream and make it readable.
func (c *Client) fetchProgress(ctx context.Context, wrapperCli *WrapperClient, ongoing *jobs, stream *jsonstream.JSONStream) error {
	var (
		ticker     = time.NewTicker(300 * time.Millisecond)
		cs         = wrapperCli.client.ContentStore()
		start      = time.Now()
		progresses = map[string]jsonstream.JSONMessage{}
		done       bool
	)
	defer ticker.Stop()

outer:
	for {
		select {
		case <-ticker.C:
			resolved := jsonstream.PullStatusResolved
			if !ongoing.isResolved() {
				resolved = jsonstream.PullStatusResolving
			}
			progresses[ongoing.name] = jsonstream.JSONMessage{
				ID:     ongoing.name,
				Status: resolved,
				Detail: &jsonstream.ProgressDetail{},
			}
			keys := []string{ongoing.name}

			activeSeen := map[string]struct{}{}
			if !done {
				actives, err := cs.ListStatuses(context.TODO(), "")
				if err != nil {
					logrus.Errorf("failed to list statuses: %v", err)
					continue
				}
				// update status of active entries!
				for _, active := range actives {
					progresses[active.Ref] = jsonstream.JSONMessage{
						ID:     active.Ref,
						Status: jsonstream.PullStatusDownloading,
						Detail: &jsonstream.ProgressDetail{
							Current: active.Offset,
							Total:   active.Total,
						},
						StartedAt: active.StartedAt,
						UpdatedAt: active.UpdatedAt,
					}
					activeSeen[active.Ref] = struct{}{}
				}
			}

			// now, update the items in jobs that are not in active
			for _, j := range ongoing.jobs() {
				key := remotes.MakeRefKey(ctx, j)
				keys = append(keys, key)
				if _, ok := activeSeen[key]; ok {
					continue
				}

				status, ok := progresses[key]
				if !done && (!ok || status.Status == jsonstream.PullStatusDownloading) {
					info, err := cs.Info(context.TODO(), j.Digest)
					if err != nil {
						if !errdefs.IsNotFound(err) {
							logrus.Errorf("failed to get content info: %v", err)
							continue outer
						} else {
							progresses[key] = jsonstream.JSONMessage{
								ID:     key,
								Status: jsonstream.PullStatusWaiting,
							}
						}
					} else if info.CreatedAt.After(start) {
						progresses[key] = jsonstream.JSONMessage{
							ID:     key,
							Status: jsonstream.PullStatusDone,
							Detail: &jsonstream.ProgressDetail{
								Current: info.Size,
								Total:   info.Size,
							},
							UpdatedAt: info.CreatedAt,
						}
					} else {
						progresses[key] = jsonstream.JSONMessage{
							ID:     key,
							Status: jsonstream.PullStatusExists,
						}
					}
				} else if done {
					if ok {
						if status.Status != jsonstream.PullStatusDone &&
							status.Status != jsonstream.PullStatusExists {

							status.Status = jsonstream.PullStatusDone
							progresses[key] = status
						}
					} else {
						progresses[key] = jsonstream.JSONMessage{
							ID:     key,
							Status: jsonstream.PullStatusDone,
						}
					}
				}
			}

			for _, key := range keys {
				stream.WriteObject(progresses[key])
			}

			if done {
				return nil
			}
		case <-ctx.Done():
			done = true // allow ui to update once more
		}
	}
}

type jobs struct {
	name     string
	added    map[digest.Digest]struct{}
	descs    []ocispec.Descriptor
	mu       sync.Mutex
	resolved bool
}

func newJobs(name string) *jobs {
	return &jobs{
		name:  name,
		added: map[digest.Digest]struct{}{},
	}
}

func (j *jobs) add(desc ocispec.Descriptor) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.resolved = true

	if _, ok := j.added[desc.Digest]; ok {
		return
	}
	j.descs = append(j.descs, desc)
	j.added[desc.Digest] = struct{}{}
}

func (j *jobs) jobs() []ocispec.Descriptor {
	j.mu.Lock()
	defer j.mu.Unlock()

	var descs []ocispec.Descriptor
	return append(descs, j.descs...)
}

func (j *jobs) isResolved() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.resolved
}
