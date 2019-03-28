package daemon

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/builder"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/mgr"
)

func (d *Daemon) runBuilderServer() error {
	// init options
	cfg := builder.Config{
		Debug: d.config.Debug,
		Root:  filepath.Join(d.config.HomeDir, "buildkit"),
	}
	cfg.ContainerdWorker.Address = d.config.ContainerdAddr
	cfg.ContainerdWorker.Namespace = d.config.DefaultNamespace
	cfg.ContainerdWorker.Snapshotter = d.config.Snapshotter

	bs, err := builder.New(&builder.Options{
		Config:              cfg,
		PostImageExportFunc: d.postBuildExporter(),
	})
	if err != nil {
		return err
	}

	return bs.Serve()
}

// postBuildExporter refreshes the image store cache and unpack it
//
// TODO(fuweid): the buildkit doesn't support unpack option right now. In
// order to run the builded image by pouch, we have to unpack after export.
// However, the builder user case is to build and push. It might use image
// in local.
//
// Will make the unpack as option for user.
func (d *Daemon) postBuildExporter() func(context.Context, map[string]string) error {
	var (
		imageService ctrd.ImageAPIClient = d.ctrdClient
		imageStore   mgr.ImageMgr        = d.imageMgr
		snapshotter                      = d.config.Snapshotter
	)

	return func(ctx context.Context, meta map[string]string) error {
		targetNames := meta["image.name"]
		if targetNames != "" {
			// TODO(fuweid): should we remove the image if refresh
			// or unpack fails.
			for _, name := range strings.Split(targetNames, ",") {
				img, err := imageService.GetImage(ctx, name)
				if err != nil {
					return err
				}

				if err := imageStore.StoreImageReference(ctx, img); err != nil {
					return err
				}

				if err := img.Unpack(ctx, snapshotter); err != nil {
					return err
				}
			}
		}
		return nil
	}
}
