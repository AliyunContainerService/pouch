package daemon

import (
	"context"
	"path"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/alibaba/pouch/apis/server"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/internal"
)

// Daemon refers to a daemon.
type Daemon struct {
	config         config.Config
	containerStore *meta.Store
	containerd     *ctrd.Client
	containerMgr   mgr.ContainerMgr
	systemMgr      mgr.SystemMgr
	imageMgr       mgr.ImageMgr
	volumeMgr      mgr.VolumeMgr
	server         server.Server
}

// router represents the router of daemon.
type router struct {
	daemon *Daemon
	*mux.Router
}

// NewDaemon constructs a brand new server.
func NewDaemon(cfg config.Config) *Daemon {
	containerStore, err := meta.NewStore(meta.Config{
		BaseDir: path.Join(cfg.HomeDir, "containers"),
		Buckets: []meta.Bucket{
			{
				Name: meta.MetaJSONFile,
				Type: reflect.TypeOf(mgr.ContainerMeta{}),
			},
		},
	})
	if err != nil {
		logrus.Errorf("failed to create container meta store: %v", err)
		return nil
	}

	containerd, err := ctrd.NewClient(ctrd.Config{
		Address: cfg.ContainerdAddr,
	})
	if err != nil {
		logrus.Errorf("failed to new containerd's client: %v", err)
		return nil
	}

	return &Daemon{
		config:         cfg,
		containerd:     containerd,
		containerStore: containerStore,
	}
}

// Run starts daemon.
func (d *Daemon) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	imageMgr, err := internal.GenImageMgr(&d.config, d)
	if err != nil {
		return err
	}
	d.imageMgr = imageMgr

	systemMgr, err := internal.GenSystemMgr(&d.config)
	if err != nil {
		return err
	}
	d.systemMgr = systemMgr

	volumeMgr, err := internal.GenVolumeMgr(&d.config, d)
	if err != nil {
		return err
	}
	d.volumeMgr = volumeMgr

	containerMgr, err := internal.GenContainerMgr(ctx, d)
	if err != nil {
		return err
	}
	d.containerMgr = containerMgr

	d.server = server.Server{
		Config:       d.config,
		ContainerMgr: containerMgr,
		SystemMgr:    systemMgr,
		ImageMgr:     imageMgr,
		VolumeMgr:    volumeMgr,
	}

	return d.server.Start()
}

// Shutdown stops daemon.
func (d *Daemon) Shutdown() error {
	return d.server.Stop()
}

// Config gets config of daemon.
func (d *Daemon) Config() *config.Config {
	return &d.config
}

// ImgMgr gets manager of image.
func (d *Daemon) ImgMgr() mgr.ImageMgr {
	return d.imageMgr
}

// VolMgr gets manager of volume.
func (d *Daemon) VolMgr() mgr.VolumeMgr {
	return d.volumeMgr
}

// Containerd gets containerd client.
func (d *Daemon) Containerd() *ctrd.Client {
	return d.containerd
}

// MetaStore gets store of meta.
func (d *Daemon) MetaStore() *meta.Store {
	return d.containerStore
}
