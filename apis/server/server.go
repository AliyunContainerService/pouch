package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/alibaba/pouch/apis/plugins"
	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"

	"github.com/sirupsen/logrus"
)

// Server is a http server which serves restful api to client.
type Server struct {
	Config           *config.Config
	ContainerMgr     mgr.ContainerMgr
	SystemMgr        mgr.SystemMgr
	ImageMgr         mgr.ImageMgr
	VolumeMgr        mgr.VolumeMgr
	NetworkMgr       mgr.NetworkMgr
	listeners        []net.Listener
	ContainerPlugin  plugins.ContainerPlugin
	ManagerWhiteList map[string]struct{}
	lock             sync.RWMutex
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
		tlsConfig, err = client.GenTLSConfig(s.Config.TLS.Key, s.Config.TLS.Cert, s.Config.TLS.CA)
		if err != nil {
			return err
		}
		if s.Config.TLS.VerifyRemote {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		SetupManagerWhitelist(s)
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

// SetupManagerWhitelist enables users to setup which common name can access this server
func SetupManagerWhitelist(server *Server) {
	if server.Config.TLS.ManagerWhiteList != "" {
		server.lock.Lock()
		defer server.lock.Unlock()
		arr := strings.Split(server.Config.TLS.ManagerWhiteList, ",")
		server.ManagerWhiteList = make(map[string]struct{}, len(arr))
		for _, cn := range arr {
			server.ManagerWhiteList[cn] = struct{}{}
		}
	}
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
