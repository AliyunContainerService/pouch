package builder

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/containerd/containerd/sys"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/worker/base"
)

func addPostImageExporter(w *base.Worker, postExportFunc func(context.Context, map[string]string) error) (*base.Worker, error) {
	exp, err := w.Exporter(client.ExporterImage)
	if err != nil {
		return nil, err
	}

	w.Exporters[client.ExporterImage] = &wrapperImageExporter{
		Exporter:       exp,
		postExportFunc: postExportFunc,
	}
	return w, nil
}

func getListener(addr string) (net.Listener, error) {
	addrSlice := strings.SplitN(addr, "://", 2)
	if len(addrSlice) < 2 {
		return nil, fmt.Errorf("address %s does not contain proto", addr)
	}

	proto := addrSlice[0]
	listenAddr := addrSlice[1]
	switch proto {
	case "unix":
		return sys.GetLocalListener(listenAddr, 0, 0)
	case "tcp":
		return net.Listen("tcp", listenAddr)
	default:
		return nil, fmt.Errorf("proto %s not supported", proto)
	}
}
