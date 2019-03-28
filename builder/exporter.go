package builder

import (
	"context"

	"github.com/moby/buildkit/exporter"
)

type wrapperImageExporter struct {
	exporter.Exporter

	postExportFunc func(context.Context, map[string]string) error
}

func (w *wrapperImageExporter) Resolve(ctx context.Context, meta map[string]string) (exporter.ExporterInstance, error) {
	i, err := w.Exporter.Resolve(ctx, meta)
	if err != nil {
		return nil, err
	}

	return &wrapperExporterInstance{
		ExporterInstance: i,
		postExportFunc:   w.postExportFunc,
	}, nil
}

type wrapperExporterInstance struct {
	exporter.ExporterInstance

	postExportFunc func(context.Context, map[string]string) error
}

func (eei *wrapperExporterInstance) Export(ctx context.Context, src exporter.Source) (map[string]string, error) {
	res, err := eei.ExporterInstance.Export(ctx, src)
	if err != nil {
		return nil, err
	}

	if eei.postExportFunc != nil {
		if err := eei.postExportFunc(ctx, res); err != nil {
			return nil, err
		}
	}
	return res, nil
}
