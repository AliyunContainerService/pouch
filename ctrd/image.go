package ctrd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/jsonstream"
	"github.com/alibaba/pouch/pkg/reference"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	ctrdmetaimages "github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/opencontainers/go-digest"
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
		res        = make([]containerd.Image, 0, len(imgs))
		snaphotter = CurrentSnapshotterName(ctx)
	)

	for _, img := range imgs {
		image := containerd.NewImage(wrapperCli.client, img)

		err = image.Unpack(ctx, snaphotter)
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

	resolver, _, err := c.ResolveImage(ctx, ref, []string{ref}, authConfig, docker.ResolverOptions{
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

// ResolveImage attempts to resolve the image reference into a available reference and resolver.
func (c *Client) ResolveImage(ctx context.Context, nameRef string, refs []string, authConfig *types.AuthConfig, opts docker.ResolverOptions) (remotes.Resolver, string, error) {
	resolver, availableRef, err := c.getResolver(ctx, authConfig, nameRef, refs, opts)
	if err != nil {
		logrus.Errorf("image ref not found %s", nameRef)
		return nil, "", err
	}

	return resolver, availableRef, nil
}

// FetchImage fetches image content from the remote repository.
func (c *Client) FetchImage(ctx context.Context, resolver remotes.Resolver, availableRef string, authConfig *types.AuthConfig, stream *jsonstream.JSONStream) (containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	ongoing := newJobs(availableRef)

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
	return img, nil
}

func (c *Client) fetchImage(ctx context.Context, wrapperCli *WrapperClient, ref string, options []containerd.RemoteOpt) (containerd.Image, error) {
	img, err := wrapperCli.client.Pull(ctx, ref, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull image")
	}

	return img, nil
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
