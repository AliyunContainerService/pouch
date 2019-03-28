package builder

import (
	"fmt"

	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/worker/base"
	workerctrd "github.com/moby/buildkit/worker/containerd"
)

type workerOpt = base.WorkerOpt

type workerOptFunc func(*workerOpt) error

func withWorkerSessionManager(mgr *session.Manager) workerOptFunc {
	return func(opt *workerOpt) error {
		if opt.SessionManager != nil {
			return fmt.Errorf("already assign the session.Manager")
		}

		opt.SessionManager = mgr
		return nil
	}
}

func initializeContainerdWorker(cfg *Config, opts ...workerOptFunc) (*base.Worker, error) {
	wopt, err := workerctrd.NewWorkerOpt(
		cfg.Root,
		cfg.ContainerdWorker.Address,
		cfg.ContainerdWorker.Snapshotter,
		cfg.ContainerdWorker.Namespace,
		nil,
	)
	if err != nil {
		return nil, err
	}

	for _, o := range opts {
		if err := o(&wopt); err != nil {
			return nil, err
		}
	}
	return base.NewWorker(wopt)
}
