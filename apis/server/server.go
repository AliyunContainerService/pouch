package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"

	"github.com/alibaba/pouch/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Server is a http server which serves restful api to client.
type Server struct {
	Config       config.Config
	ContainerMgr mgr.ContainerMgr
	SystemMgr    mgr.SystemMgr
	NetworkMgr	 mgr.NetworkMgr
	ImageMgr     mgr.ImageMgr
	VolumeMgr    mgr.VolumeMgr
	listeners    []net.Listener
}

// Start setup route table and listen to specified address which currently only supports unix socket and tcp address.
func (s *Server) Start() (err error) {
	router := initRoute(s)
	errCh := make(chan error)

	defer func() {
		if err != nil {
			for _, one := range s.listeners {
				one.Close()
			}
		}
	}()

	var tlsConfig *tls.Config
	if s.Config.TLS.Key != "" && s.Config.TLS.Cert != "" {
		tlsConfig, err = utils.GenTLSConfig(s.Config.TLS.Key, s.Config.TLS.Cert, s.Config.TLS.CA)
		if err != nil {
			return err
		}
		if s.Config.TLS.VerifyRemote {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	for _, one := range s.Config.Listen {
		l, err := getListener(one, tlsConfig)
		if err != nil {
			return err
		}
		logrus.Infof("start to listen to: %s", one)
		s.listeners = append(s.listeners, l)

		go func(l net.Listener) {
			errCh <- http.Serve(l, router)
		}(l)
	}

	// not error, will block and run forever.
	return <-errCh
}

// Stop will shutdown http server by closing all listeners.
func (s *Server) Stop() error {
	for _, one := range s.listeners {
		one.Close()
	}
	return nil
}

func getListener(addr string, tlsConfig *tls.Config) (net.Listener, error) {
	addrParts := strings.SplitN(addr, "://", 2)
	if len(addrParts) != 2 {
		return nil, fmt.Errorf("invalid listening address: %s", addr)
	}

	switch addrParts[0] {
	case "tcp":
		l, err := net.Listen("tcp", addrParts[1])
		if err != nil {
			return l, err
		}
		if tlsConfig != nil {
			l = tls.NewListener(l, tlsConfig)
		}
		return l, err
	case "unix":
		if err := syscall.Unlink(addrParts[1]); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		mask := syscall.Umask(0777)
		defer syscall.Umask(mask)
		return net.Listen("unix", addrParts[1])
	default:
		return nil, fmt.Errorf("only unix socket or tcp address is support")
	}
}
