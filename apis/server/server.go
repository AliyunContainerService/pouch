package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"
)

// Server is a http server which serves restful api to client.
type Server struct {
	Config       config.Config
	ContainerMgr mgr.ContainerMgr
	SystemMgr    mgr.SystemMgr
	ImageMgr     mgr.ImageMgr
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

	for _, one := range s.Config.Listen {
		l, err := getListener(one)
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

func getListener(addr string) (net.Listener, error) {
	addrParts := strings.SplitN(addr, "://", 2)
	if len(addrParts) != 2 {
		return nil, fmt.Errorf("Invalid listen address: %s", addr)
	}

	switch addrParts[0] {
	case "tcp":
		return net.Listen("tcp", addrParts[1])
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
