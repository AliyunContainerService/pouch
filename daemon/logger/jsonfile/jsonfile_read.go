package jsonfile

import (
	"os"

	"github.com/alibaba/pouch/daemon/logger"
)

// ReadLogMessages will create goroutine to read the log message and send it to
// LogWatcher.
func (lf *JSONLogFile) ReadLogMessages(cfg *logger.ReadConfig) *logger.LogWatcher {
	watcher := logger.NewLogWatcher()

	go func() {
		// NOTE: We cannot close the channel in the JSONLogFile.read
		// function because we cannot guarantee that watcher will be
		// close the channel. Since the watcher is created in the
		// JSONLogFile.ReadLogMessages, we make sure that watcher.Msgs
		// can be closed after the JSONLogFile.read.
		defer close(watcher.Msgs)

		lf.read(cfg, watcher)
	}()
	return watcher
}

func (lf *JSONLogFile) read(cfg *logger.ReadConfig, watcher *logger.LogWatcher) {
	lf.mu.Lock()
	f, err := os.Open(lf.f.Name())
	lf.mu.Unlock()

	if err != nil {
		watcher.Err <- err
		return
	}
	defer f.Close()

	// find the offset if the config contains the valid tail lines
	if cfg.Tail > 0 {
		offset, err := seekOffsetByTailLines(f, cfg.Tail)
		if err != nil {
			watcher.Err <- err
			return
		}

		if _, err := f.Seek(offset, os.SEEK_SET); err != nil {
			watcher.Err <- err
			return
		}
	}
	tailFile(f, cfg, newUnmarshal, watcher)

	if !cfg.Follow {
		return
	}

	followFile(f, cfg, newUnmarshal, watcher)
}
