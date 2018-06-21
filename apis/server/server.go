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
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/user"

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
		tlsConfig, err = httputils.GenTLSConfig(s.Config.TLS.Key, s.Config.TLS.Cert, s.Config.TLS.CA)
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
		return newUnixSocket(addrParts[1])

	default:
		return nil, fmt.Errorf("only unix socket or tcp address is support")
	}
}

func newUnixSocket(path string) (net.Listener, error) {
	if err := syscall.Unlink(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	oldmask := syscall.Umask(0777)
	defer syscall.Umask(oldmask)
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	// chmod unix socket, make other group writable
	if err := os.Chmod(path, 0660); err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to chmod %s: %s", path, err)
	}

	gid, err := user.ParseID(user.GroupFile, "pouch", func(line, str string, idInt int, idErr error) (uint32, bool) {
		var (
			name, placeholder string
			id                int
		)

		user.ParseString(line, &name, &placeholder, &id)
		if str == name {
			return uint32(id), true
		}
		return 0, false
	})
	if err != nil {
		// ignore error when group pouch not exist, group pouch should to be
		// created before pouchd started, it means code not create pouch group
		logrus.Warnf("failed to find group pouch, cannot change unix socket %s to pouch group", path)
		return l, nil
	}

	// chown unix socket with group pouch
	if err := os.Chown(path, 0, int(gid)); err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to chown %s: %s", path, err)
	}
	return l, nil
}
