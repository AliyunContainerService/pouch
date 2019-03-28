package builder

import (
	"context"
	"os"
	"path/filepath"

	"github.com/moby/buildkit/control"
	"github.com/moby/buildkit/frontend"
	dockerfile "github.com/moby/buildkit/frontend/dockerfile/builder"
	"github.com/moby/buildkit/frontend/gateway"
	"github.com/moby/buildkit/frontend/gateway/forwarder"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/solver/bboltcachestorage"
	"github.com/moby/buildkit/worker"
	"google.golang.org/grpc"
)

// Options is used to config the BuilderServer.
type Options struct {
	Config Config

	// PostImageExportFunc allows to do post check or post action
	// after exporter.
	//
	// TODO(fuweid): Since we caches containerd image in our own, pouch
	// images command won't get the new image exported by buildkit if
	// we don't store it in the cache. The post check is used to
	// save the cache in local. I think we should remove this option
	// after we remove the cache.
	PostImageExportFunc func(context.Context, map[string]string) error
}

// Server wrappers buildkit to provide builder functionality.
type Server struct {
	cfg        *Config
	srv        *grpc.Server
	controller *control.Controller
}

// New returns Server.
//
// TODO(fuweid):
// 1. supports network mode in containerd worker.
// 2. supports cpu/memory limitation in containerd worker.
// 3. supports registry cache.
// 4. use runC/PouchContainer's container mgr to run the container
//	container for build is not created by PouchContainer's ContainerMgr
//	and the exit event will be filed to PouchContainer which has no idea
//	about this. The error log will be annoying.
func New(opts *Options) (*Server, error) {
	sessionMgr, err := session.NewManager()
	if err != nil {
		return nil, err
	}

	setDefaultConfig(&opts.Config)

	// create root
	if err := os.MkdirAll(opts.Config.Root, 0700); err != nil {
		return nil, err
	}

	// initialize containerd worker
	w, err := initializeContainerdWorker(
		&opts.Config,
		withWorkerSessionManager(sessionMgr),
	)
	if err != nil {
		return nil, err
	}

	w, err = addPostImageExporter(w, opts.PostImageExportFunc)
	if err != nil {
		return nil, err
	}

	wc := &worker.Controller{}
	if err := wc.Add(w); err != nil {
		return nil, err
	}

	// frontends
	frontends := map[string]frontend.Frontend{}
	frontends["dockerfile.v0"] = forwarder.NewGatewayForwarder(wc, dockerfile.Build)
	frontends["gateway.v0"] = gateway.NewGatewayFrontend(wc)

	// cacheStorage
	cacheStorage, err := bboltcachestorage.NewStore(filepath.Join(opts.Config.Root, "cache.db"))
	if err != nil {
		return nil, err
	}

	// generate controller
	ctrl, err := control.NewController(control.Opt{
		Frontends:        frontends,
		SessionManager:   sessionMgr,
		CacheKeyStorage:  cacheStorage,
		WorkerController: wc,
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:        &opts.Config,
		controller: ctrl,
	}, nil
}

// Serve starts the Server.
func (bs *Server) Serve() error {
	srv := grpc.NewServer()
	bs.controller.Register(srv)

	// listener
	lis, err := getListener(bs.cfg.GRPC.Address)
	if err != nil {
		return err
	}
	defer lis.Close()
	bs.srv = srv

	return bs.srv.Serve(lis)
}

// Stop the Server.
func (bs *Server) Stop() {
	if bs.srv == nil {
		return
	}

	srv := bs.srv
	bs.srv = nil
	srv.GracefulStop()
}
