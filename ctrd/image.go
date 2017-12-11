package ctrd

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/jsonstream"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// RemoveImage deletes an image.
func (c *Client) RemoveImage(ctx context.Context, ref string) error {
	err := c.client.ImageService().Delete(ctx, ref)
	if err != nil {
		return errors.Wrap(err, "failed to remove image")
	}
	return nil
}

// ListImages lists all images.
func (c *Client) ListImages(ctx context.Context, filter ...string) ([]types.ImageInfo, error) {
	imageList, err := c.client.ImageService().List(ctx, filter...)
	if err != nil {
		return nil, err
	}

	images := make([]types.ImageInfo, 0, 32)
	digestPrefix := "sha256:"
	for _, image := range imageList {
		descriptor := image.Target
		digest := []byte(descriptor.Digest)

		size, err := image.Size(ctx, c.client.ContentStore(), platforms.Default())
		if err != nil {
			return nil, err
		}

		images = append(images, types.ImageInfo{
			Name:   image.Name,
			ID:     string(digest[len(digestPrefix) : len(digestPrefix)+12]),
			Digest: string(digest),
			Size:   size,
		})
	}
	return images, nil
}

// PullImage downloads an image from the remote repository.
func (c *Client) PullImage(ctx context.Context, ref string, stream *jsonstream.JSONStream) (containerd.Image, error) {
	resolver, err := resolver()
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
		if desc.MediaType != images.MediaTypeDockerSchema1Manifest {
			ongoing.add(desc)
		}
		return nil, nil
	}
	options = append(options, containerd.WithImageHandler(images.HandlerFunc(handle)))

	// fetch progress status, then send to client via out channel.
	pctx, cancelProgress := context.WithCancel(ctx)
	wait := make(chan struct{})

	go func() {
		if err := c.fetchProgress(pctx, ongoing, stream); err != nil {
			logrus.Errorf("failed to get pull's progress: %v", err)
		}
		close(wait)

		logrus.Infof("fetch progress exited, ref: %s.", ref)
	}()

	// start to pull image.
	img, err := c.pullImage(ctx, ref, options)

	// cancel fetch progress befor handle error.
	cancelProgress()
	defer stream.Close()

	// wait fetch progress to finish.
	<-wait

	if err != nil {
		// Send Error information to client through stream
		messages := []ProgressInfo{
			{Code: http.StatusInternalServerError, ErrorMessage: err.Error()},
		}
		stream.WriteObject(messages)

		return nil, err
	}

	logrus.Infof("success to pull image: %s", img.Name())
	return img, nil
}

func (c *Client) pullImage(ctx context.Context, ref string, options []containerd.RemoteOpt) (containerd.Image, error) {
	ctx = leases.WithLease(ctx, c.lease.ID())

	img, err := c.client.Pull(ctx, ref, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull image")
	}

	return img, nil
}

// ProgressInfo represents the status of downloading image.
type ProgressInfo struct {
	Ref       string
	Status    string
	Offset    int64
	Total     int64
	StartedAt time.Time
	UpdatedAt time.Time

	// For Error handling
	Code         int    // http response code
	ErrorMessage string // detail error information
}

func (c *Client) fetchProgress(ctx context.Context, ongoing *jobs, stream *jsonstream.JSONStream) error {
	var (
		ticker     = time.NewTicker(100 * time.Millisecond)
		cs         = c.client.ContentStore()
		start      = time.Now()
		progresses = map[string]ProgressInfo{}
		done       bool
	)
	defer ticker.Stop()

outer:
	for {
		select {
		case <-ticker.C:
			resolved := "resolved"
			if !ongoing.isResolved() {
				resolved = "resolving"
			}
			progresses[ongoing.name] = ProgressInfo{
				Ref:    ongoing.name,
				Status: resolved,
			}
			keys := []string{ongoing.name}

			activeSeen := map[string]struct{}{}
			if !done {
				active, err := cs.ListStatuses(context.TODO(), "")
				if err != nil {
					logrus.Errorf("failed to list statuses: %v", err)
					continue
				}
				// update status of active entries!
				for _, active := range active {
					progresses[active.Ref] = ProgressInfo{
						Ref:       active.Ref,
						Status:    "downloading",
						Offset:    active.Offset,
						Total:     active.Total,
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
				if !done && (!ok || status.Status == "downloading") {
					info, err := cs.Info(context.TODO(), j.Digest)
					if err != nil {
						if !errdefs.IsNotFound(err) {
							logrus.Errorf("failed to get content info: %v", err)
							continue outer
						} else {
							progresses[key] = ProgressInfo{
								Ref:    key,
								Status: "waiting",
							}
						}
					} else if info.CreatedAt.After(start) {
						progresses[key] = ProgressInfo{
							Ref:       key,
							Status:    "done",
							Offset:    info.Size,
							Total:     info.Size,
							UpdatedAt: info.CreatedAt,
						}
					} else {
						progresses[key] = ProgressInfo{
							Ref:    key,
							Status: "exists",
						}
					}
				} else if done {
					if ok {
						if status.Status != "done" && status.Status != "exists" {
							status.Status = "done"
							progresses[key] = status
						}
					} else {
						progresses[key] = ProgressInfo{
							Ref:    key,
							Status: "done",
						}
					}
				}
			}

			var ordered []ProgressInfo
			for _, key := range keys {
				ordered = append(ordered, progresses[key])
			}

			stream.WriteObject(ordered)

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
