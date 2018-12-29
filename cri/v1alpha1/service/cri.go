package service

import (
	"github.com/alibaba/pouch/cri/middleware"
	cri "github.com/alibaba/pouch/cri/v1alpha1"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/netutils"

	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// Service serves the kubelet runtime grpc api which will be consumed by kubelet.
type Service struct {
	config *config.Config
	server *grpc.Server
	criMgr cri.CriMgr
}

// NewService creates a brand new cri service.
func NewService(cfg *config.Config, criMgr cri.CriMgr) (*Service, error) {
	s := &Service{
		config: cfg,
		server: grpc.NewServer(
			grpc.UnaryInterceptor(middleware.HandleWithGlobalMiddlewares(nil)),
		),
		criMgr: criMgr,
	}

	runtime.RegisterRuntimeServiceServer(s.server, s.criMgr)
	runtime.RegisterImageServiceServer(s.server, s.criMgr)

	return s, nil
}

// Serve starts grpc server.
func (s *Service) Serve() error {
	l, err := netutils.GetListener(s.config.CriConfig.Listen, nil)
	if err != nil {
		return err
	}

	return s.server.Serve(l)
}
