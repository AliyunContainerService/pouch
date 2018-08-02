package v1alpha1

import (
	"fmt"
	"net"

	"github.com/alibaba/pouch/cri/stream"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/netutils"
)

func newStreamServer(ctrMgr mgr.ContainerMgr, address string, port string) (Server, error) {
	if address == "" {
		a, err := netutils.ChooseBindAddress(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get stream server address: %v", err)
		}
		address = a.String()
	}
	config := stream.DefaultConfig
	config.Address = net.JoinHostPort(address, port)
	runtime := stream.NewStreamRuntime(ctrMgr)
	return NewServer(config, runtime)
}
