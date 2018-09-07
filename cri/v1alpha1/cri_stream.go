package v1alpha1

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/alibaba/pouch/cri/stream"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/netutils"

	"github.com/sirupsen/logrus"
)

func newStreamServer(ctrMgr mgr.ContainerMgr, address string, port string, reuseHTTPSPort bool) (Server, error) {
	ip := net.ParseIP(address)
	// If the address is "" or "0.0.0.0", choose a proper one by ourselves.
	if ip == nil || ip.IsUnspecified() {
		a, err := netutils.ChooseBindAddress(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get stream server address: %v", err)
		}
		address = a.String()
	}
	config := stream.DefaultConfig
	config.Address = net.JoinHostPort(address, port)
	config.BaseURL = &url.URL{
		Scheme: "http",
		Host:   config.Address,
	}
	if reuseHTTPSPort {
		config.BaseURL.Scheme = "https"
	}
	logrus.Infof("Stream Server will bind to address %v", config.Address)

	runtime := stream.NewStreamRuntime(ctrMgr)
	return NewServer(config, runtime)
}

// extractIPAndPortFromAddresses extract first valid ip and port from addresses.
func extractIPAndPortFromAddresses(addresses []string) (string, string) {
	for _, addr := range addresses {
		addrParts := strings.SplitN(addr, "://", 2)
		if len(addrParts) != 2 {
			logrus.Errorf("invalid listening address %s: must be in format [protocol]://[address]", addr)
			continue
		}

		switch addrParts[0] {
		case "tcp":
			host, port, err := net.SplitHostPort(addrParts[1])
			if err != nil {
				logrus.Errorf("failed to split host and port from address: %v", err)
				continue
			}
			return host, port
		case "unix":
			continue
		default:
			logrus.Errorf("only unix socket or tcp address is support")
		}
	}
	return "", ""
}
