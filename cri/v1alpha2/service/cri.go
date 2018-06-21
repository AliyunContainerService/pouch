package service

import (
	"net"
	"os"
	"syscall"

	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
	cri "github.com/alibaba/pouch/cri/v1alpha2"
	"github.com/alibaba/pouch/daemon/config"

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

	// TODO: Prepare streaming server.

	runtime.RegisterRuntimeServiceServer(s.server, s.criMgr)
	runtime.RegisterImageServiceServer(s.server, s.criMgr)

	return s, nil
}

// Serve starts grpc server.
func (s *Service) Serve() error {
	// Unlink to cleanup the previous socket file.
	if err := syscall.Unlink(s.config.CriConfig.Listen); err != nil && !os.IsNotExist(err) {
		return err
	}

	l, err := net.Listen("unix", s.config.CriConfig.Listen)
	if err != nil {
		return err
	}

	return s.server.Serve(l)
}
