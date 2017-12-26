package daemon

import (
	"context"
	"path"
	"reflect"

	"github.com/alibaba/pouch/apis/server"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/internal"
	"github.com/alibaba/pouch/network/bridge"
	"github.com/alibaba/pouch/cri"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
	networkMgr     mgr.NetworkMgr
	criManager 	   *cri.CRIManager
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

	networkMgr, err := internal.GenNetworkMgr(&d.config, d)
	if err != nil {
		return err
	}
	d.networkMgr = networkMgr

	containerMgr, err := internal.GenContainerMgr(ctx, d)
	if err != nil {
		return err
	}
	d.containerMgr = containerMgr

	criManager, err := internal.GenCRIManager(d)
	if err != nil {
		return err
	}
	d.criManager = criManager

	d.server = server.Server{
		Config:       d.config,
		ContainerMgr: containerMgr,
		SystemMgr:    systemMgr,
		ImageMgr:     imageMgr,
		VolumeMgr:    volumeMgr,
		NetworkMgr:   networkMgr,
	}

	// init base network
	err = d.networkInit(ctx)
	if err != nil {
		return err
	}

	httpServerCloseCh := make(chan struct{})
	go func() {
		if err := d.server.Start(); err != nil {
			logrus.Errorf("Failed to start http server: %v", err)
		}
		close(httpServerCloseCh)
	}()

	grpcServerCloseCh := make(chan struct{})
	go func() {
		if err := d.criManager.Serve(); err != nil {
			logrus.Errorf("Failed to start grpc server: %v", err)
		}
		close(grpcServerCloseCh)
	}()

	// Stop pouchd if both server stopped.
	<-httpServerCloseCh
	logrus.Infof("HTTP server stopped")
	<-grpcServerCloseCh
	logrus.Infof("GRPC server stopped")
	return nil
}

// Shutdown stops daemon.
func (d *Daemon) Shutdown() error {
	return d.server.Stop()
}

// Config gets config of daemon.
func (d *Daemon) Config() *config.Config {
	return &d.config
}

// CtrMgr gets manager of container.
func (d *Daemon) CtrMgr() mgr.ContainerMgr {
	return d.containerMgr
}

// ImgMgr gets manager of image.
func (d *Daemon) ImgMgr() mgr.ImageMgr {
	return d.imageMgr
}

// VolMgr gets manager of volume.
func (d *Daemon) VolMgr() mgr.VolumeMgr {
	return d.volumeMgr
}

// NetMgr gets manager of network.
func (d *Daemon) NetMgr() mgr.NetworkMgr {
	return d.networkMgr
}

// Containerd gets containerd client.
func (d *Daemon) Containerd() *ctrd.Client {
	return d.containerd
}

// MetaStore gets store of meta.
func (d *Daemon) MetaStore() *meta.Store {
	return d.containerStore
}

func (d *Daemon) networkInit(ctx context.Context) error {
	return bridge.New(ctx, d.config.NetworkConfg.BridgeConfig, d.networkMgr)
}
