package ctrd

import (
	"context"
	"fmt"
	"sync"

	"github.com/containerd/containerd"
	"github.com/pkg/errors"
)

// WrapperClient wrappers containerd grpc client,
// so that pouch daemon can holds a grpc client pool
// to improve grpc client performance.
type WrapperClient struct {
	client *containerd.Client

	// Lease is a new feature of containerd, We use it to avoid that the images
	// are removed by garbage collection. If no lease is defined, the downloaded images will
	// be removed automatically when the container is removed.
	lease *containerd.Lease

	mux sync.Mutex
	// streamQuota records the numbers of stream client without be using
	streamQuota int
}

func newWrapperClient(rpcAddr string, maxStreamsClient int) (*WrapperClient, error) {
	options := []containerd.ClientOpt{
		containerd.WithDefaultNamespace("default"),
	}
	cli, err := containerd.New(rpcAddr, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect containerd")
	}

	// create a new lease or reuse the existed.
	var lease containerd.Lease

	leases, err := cli.ListLeases(context.TODO())
	if err != nil {
		return nil, err
	}
	if len(leases) != 0 {
		lease = leases[0]
	} else {
		if lease, err = cli.CreateLease(context.TODO()); err != nil {
			return nil, err
		}
	}

	return &WrapperClient{
		client:      cli,
		lease:       &lease,
		streamQuota: maxStreamsClient,
	}, nil
}

// Produce is to release specified numbers of grpc stream client
// FIXME(ziren): if streamQuota greater than defaultMaxStreamsClient
// what to do ???
func (w *WrapperClient) Produce(v int) {
	w.mux.Lock()
	defer w.mux.Unlock()
	w.streamQuota += v
}

// Consume is to acquire specified numbers of grpc stream client
func (w *WrapperClient) Consume(v int) error {
	w.mux.Lock()
	defer w.mux.Unlock()

	if w.streamQuota < v {
		return fmt.Errorf("quota is %d, less than %d, can not acquire", w.streamQuota, v)
	}

	w.streamQuota -= v
	return nil
}

// Value is to get the quota
func (w *WrapperClient) Value() int {
	w.mux.Lock()
	defer w.mux.Unlock()

	return w.streamQuota
}
