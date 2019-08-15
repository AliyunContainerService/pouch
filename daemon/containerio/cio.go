package containerio

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alibaba/pouch/pkg/log"
	"github.com/containerd/containerd/cio"
)

// NewFIFOSet prepares fifo files.
func NewFIFOSet(processID string, withStdin bool, withTerminal bool) (*cio.FIFOSet, error) {
	root := "/run/containerd/fifo"
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}

	fifoDir, err := ioutil.TempDir(root, "")
	if err != nil {
		return nil, err
	}

	cfg := cio.Config{
		Terminal: withTerminal,
		Stdout:   filepath.Join(fifoDir, processID+"-stdout"),
	}

	if withStdin {
		cfg.Stdin = filepath.Join(fifoDir, processID+"-stdin")
	}

	if !withTerminal {
		cfg.Stderr = filepath.Join(fifoDir, processID+"-stderr")
	}

	closeFn := func() error {
		err := os.RemoveAll(fifoDir)
		if err != nil {
			log.With(nil).WithError(err).Warnf("failed to remove process(id=%v) fifo dir", processID)
		}
		return err
	}

	return cio.NewFIFOSet(cfg, closeFn), nil
}
