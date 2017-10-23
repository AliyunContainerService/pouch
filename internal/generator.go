package internal

import (
	"context"
	"sync"

	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/collect"
)

var (
	contMgr  *mgr.ContainerManager
	contOnce sync.Once
	contErr  error
	imgMgr   *mgr.ImageManager
	imgOnce  sync.Once
	imgErr   error
	sysMgr   *mgr.SystemManager
	sysOnce  sync.Once
	sysErr   error
)

// DaemonProvider provides resources which are needed by container manager and are from daemon.
type DaemonProvider interface {
	Config() *config.Config
	Containerd() *ctrd.Client
	MetaStore() *meta.Store
}

// GenContainerMgr generates a ContainerMgr instance according to config cfg.
func GenContainerMgr(ctx context.Context, d DaemonProvider) (mgr.ContainerMgr, error) {
	contOnce.Do(func() {
		imgMgr, err := GenImageMgr(d.Config(), d)
		if err != nil {
			contErr = err
			return
		}

		contMgr = &mgr.ContainerManager{
			Store:    d.MetaStore(),
			NameToID: collect.NewSafeMap(),
			Client:   d.Containerd(),
			ImageMgr: imgMgr,
		}

		if err := contMgr.Restore(ctx); err != nil {
			contErr = err
		}
	})

	return contMgr, contErr
}

// GenSystemMgr generates a SystemMgr instance according to config cfg.
func GenSystemMgr(cfg config.Config) (mgr.SystemMgr, error) {
	//TODO
	sysOnce.Do(func() {
		sysMgr, sysErr = mgr.NewSystemManager(cfg)
	})

	return sysMgr, sysErr
}

// GenImageMgr generates a SystemMgr instance according to config cfg.
func GenImageMgr(cfg *config.Config, d DaemonProvider) (mgr.ImageMgr, error) {
	imgOnce.Do(func() {
		imgMgr, imgErr = mgr.NewImageManager(cfg, d.Containerd())
	})

	return imgMgr, imgErr
}
