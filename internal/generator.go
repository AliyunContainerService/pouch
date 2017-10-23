package internal

import (
	"context"

	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/collect"
)

// DaemonProvider provides resources which are needed by container manager and are from daemon.
type DaemonProvider interface {
	Config() *config.Config
	ImgMgr() mgr.ImageMgr
	Containerd() *ctrd.Client
	MetaStore() *meta.Store
}

// GenContainerMgr generates a ContainerMgr instance according to config cfg.
func GenContainerMgr(ctx context.Context, d DaemonProvider) (mgr.ContainerMgr, error) {
	cm := &mgr.ContainerManager{
		Store:    d.MetaStore(),
		NameToID: collect.NewSafeMap(),
		Client:   d.Containerd(),
	}
	if err := cm.Restore(ctx); err != nil {
		return nil, err
	}
	return cm, nil
}

// GenSystemMgr generates a SystemMgr instance according to config cfg.
func GenSystemMgr(cfg config.Config) (mgr.SystemMgr, error) {
	//TODO
	return mgr.NewSystemManager(cfg)
}

// GenImageMgr generates a SystemMgr instance according to config cfg.
func GenImageMgr(cfg config.Config, d DaemonProvider) (mgr.ImageMgr, error) {
	return mgr.NewImageManager(cfg, d.Containerd())
}
