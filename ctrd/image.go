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
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CreateImageReference creates the image in the meta data in the containerd.
func (c *Client) CreateImageReference(ctx context.Context, img ctrdmetaimages.Image) (ctrdmetaimages.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return ctrdmetaimages.Image{}, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	return wrapperCli.client.ImageService().Create(ctx, img)
}

// GetImage returns the containerd's Image.
func (c *Client) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	return wrapperCli.client.GetImage(ctx, ref)
}

// ListImages lists all images.
func (c *Client) ListImages(ctx context.Context, filter ...string) ([]containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	return wrapperCli.client.ListImages(ctx, filter...)
}

// RemoveImage deletes an image.
func (c *Client) RemoveImage(ctx context.Context, ref string) error {
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

	exportedStream, err := wrapperCli.client.Export(ctx, exporter, desc)
	if err != nil {
		return nil, err
	}

	return exportedStream, nil
}

// ImportImage creates a set of images by tarstream.
//
// NOTE: One tar may have several manifests.
func (c *Client) ImportImage(ctx context.Context, importer ctrdmetaimages.Importer, reader io.Reader) ([]containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	// NOTE: The import will store the data into boltdb. But the unpack may
	// fail. It is not transaction.
	imgs, err := wrapperCli.client.Import(ctx, importer, reader)
	if err != nil {
		return nil, err
	}

	for _, img := range imgs {
		err = img.Unpack(ctx, containerd.DefaultSnapshotter)
		if err != nil {
			return nil, err
		}
	}
	return imgs, nil
}

// PullImage downloads an image from the remote repository.
func (c *Client) PullImage(ctx context.Context, ref string, authConfig *types.AuthConfig, stream *jsonstream.JSONStream) (containerd.Image, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	resolver, err := resolver(authConfig)
	if err != nil {
		return nil, err
	}

	ongoing := newJobs(ref)

	options := []containerd.RemoteOpt{
		containerd.WithPullUnpack,
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

		logrus.Infof("fetch progress exited, ref: %s.", ref)
	}()

	// start to pull image.
	img, err := c.pullImage(ctx, wrapperCli, ref, options)

	// cancel fetch progress before handle error.
	cancelProgress()
	defer stream.Close()

	// wait fetch progress to finish.
	<-wait

	if err != nil {
		// Send Error information to client through stream
		message := jsonstream.JSONMessage{
			Error: &jsonstream.JSONError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
		}
		stream.WriteObject(message)
		return nil, err
	}

	logrus.Infof("success to pull image: %s", img.Name())
	return img, nil
}

func (c *Client) pullImage(ctx context.Context, wrapperCli *WrapperClient, ref string, options []containerd.RemoteOpt) (containerd.Image, error) {
	ctx = leases.WithLease(ctx, wrapperCli.lease.ID())

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
