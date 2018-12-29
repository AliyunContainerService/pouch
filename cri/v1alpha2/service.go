package v1alpha2

import (
	"context"
	"path"

	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
	"github.com/alibaba/pouch/cri/metrics"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/grpc/interceptor"
	"github.com/alibaba/pouch/pkg/netutils"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Service serves the kubelet runtime grpc api which will be consumed by kubelet.
type Service struct {
	config *config.Config
	server *grpc.Server
}

// NewService creates a brand new cri service.
func NewService(cfg *config.Config, criMgr CriMgr) (*Service, error) {
	logEntry := logrus.NewEntry(logrus.StandardLogger())

	s := &Service{
		config: cfg,
		server: grpc.NewServer(
			grpc.StreamInterceptor(metrics.GRPCMetrics.StreamServerInterceptor()),
			grpc_middleware.WithUnaryServerChain(
				metrics.GRPCMetrics.UnaryServerInterceptor(),
				interceptor.PayloadUnaryServerInterceptor(logEntry, criLogLevelDecider),
			),
		),
	}

	runtime.RegisterRuntimeServiceServer(s.server, criMgr)
	runtime.RegisterImageServiceServer(s.server, criMgr)
	runtime.RegisterVolumeServiceServer(s.server, criMgr)

	// EnableHandlingTimeHistogram turns on recording of handling time
	// of RPCs. Histogram metrics can be very expensive for Prometheus
	// to retain and query.
	metrics.GRPCMetrics.EnableHandlingTimeHistogram()
	// Initialize all metrics.
	metrics.GRPCMetrics.InitializeMetrics(s.server)

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

func criLogLevelDecider(ctx context.Context, fullMethodName string, servingObject interface{}) logrus.Level {
	// extract methodName from fullMethodName
	// eg. extract 'StartContainer' from '/runtime.v1alpha2.RuntimeService/StartContainer'
	methodName := path.Base(fullMethodName)

	// method->logLevel map
	switch methodName {
	case // readonly methods
		"Version",
		"PodSandboxStatus",
		"ListPodSandbox",
		"ListContainers",
		"ContainerStatus",
		"ContainerStats",
		"ListContainerStats",
		"Status",
		"ListImages",
		"ImageStatus",
		"ImageFsInfo":
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}
