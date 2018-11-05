package server

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pouch/cri/stream"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/hookplugins"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/netutils"

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
	StreamRouter     stream.Router
	listeners        []net.Listener
	ContainerPlugin  hookplugins.ContainerPlugin
	APIPlugin        hookplugins.APIPlugin
	ManagerWhiteList map[string]struct{}
	lock             sync.RWMutex
}

// Start setup route table and listen to specified address which currently only supports unix socket and tcp address.
func (s *Server) Start(readyCh chan bool) (err error) {
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
		tlsConfig, err = httputils.GenTLSConfig(s.Config.TLS.Key, s.Config.TLS.Cert, s.Config.TLS.CA)
		if err != nil {
			readyCh <- false
			return err
		}
		if s.Config.TLS.VerifyRemote {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		SetupManagerWhitelist(s)
	}

	for _, one := range s.Config.Listen {
		l, err := netutils.GetListener(one, tlsConfig)
		if err != nil {
			readyCh <- false
			return err
		}
		logrus.Infof("start to listen to: %s", one)
		s.listeners = append(s.listeners, l)

		go func(l net.Listener) {
			s := &http.Server{
				Handler:           router,
				ErrorLog:          log.New(stdFilterLogWriter, "", 0),
				ReadTimeout:       time.Minute * 10,
				ReadHeaderTimeout: time.Minute * 10,
				IdleTimeout:       time.Minute * 10,
			}
			errCh <- s.Serve(l)
		}(l)
	}

	// the http server has set up, send Ready
	readyCh <- true

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
