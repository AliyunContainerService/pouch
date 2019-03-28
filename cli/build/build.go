package build

import (
	"context"
	"os"

	"github.com/containerd/console"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// Build connects to BuilderServer and build.
//
// TODO(fuweid):
// 1. support auth for private repo
func Build(ctx context.Context, addr string, opt *Options) error {
	cli, err := client.New(ctx, addr)
	if err != nil {
		return err
	}

	frontendAttrs, err := optsToFrontendAttrs(opt)
	if err != nil {
		return err
	}

	exporterAttrs, err := optsToExporterAttrs(opt)
	if err != nil {
		return err
	}

	solveOpt := client.SolveOpt{
		Exporter:      "image",
		ExporterAttrs: exporterAttrs,
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		// TODO: basically, we only need to support one workdir
		LocalDirs: opt.LocalDirs,
	}

	ch := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, err := cli.Solve(ctx, nil, solveOpt, ch)
		return err
	})

	eg.Go(func() error {
		var c console.Console

		cf, err := console.ConsoleFromFile(os.Stderr)
		if err == nil {
			c = cf
		} else {
			logrus.Debug("failed to use tty for status: %v", err)
		}

		return progressui.DisplaySolveStatus(ctx, "", c, os.Stdout, ch)
	})
	return eg.Wait()
}
