package service

import (
	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
	cri "github.com/alibaba/pouch/cri/v1alpha2"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/netutils"

	"google.golang.org/grpc"
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
		server: grpc.NewServer(),
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
