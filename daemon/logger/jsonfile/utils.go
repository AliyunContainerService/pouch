package jsonfile

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/alibaba/pouch/daemon/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type newUnmarshalFunc func(r io.Reader) func() (*logger.LogMessage, error)

var watchFileTimeout = 200 * time.Millisecond

// followFile will act like `tail -f`.
func followFile(f *os.File, cfg *logger.ReadConfig, unmarshaler newUnmarshalFunc, watcher *logger.LogWatcher) {
	fileWatcher, err := watchFileChange(f.Name())
	if err != nil {
		watcher.Err <- err
		return
	}

	defer func() {
		fileWatcher.Remove(f.Name())
		fileWatcher.Close()
	}()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-watcher.WatchClose():
			cancel()
		}
	}()

	decodeOneLine := unmarshaler(f)

	errDone := errors.New("done")

	// NOTE: avoid to use time.After in select. We need local-global timeout
	watchTimeout := time.NewTimer(time.Second)
	defer watchTimeout.Stop()

	// handleError will watch the file if the err is io.EOF so that
	// the loop can continue to read the file. Or just return the error.
	handleError := func(err error) error {
		if err != io.EOF {
			return err
		}

		for {
			watchTimeout.Reset(watchFileTimeout)

			select {
			case <-ctx.Done():
				// caused by the caller and we should return done
				return errDone
			case e := <-fileWatcher.Events:
				switch e.Op {
				case fsnotify.Write:
					decodeOneLine = unmarshaler(f)
					return nil
				case fsnotify.Remove:
					// ideally, it's caused by removing the container.
					return errDone
				default:
					logrus.Debug("unexpected file change during watching file %v: %v", f.Name(), e.Op)
					return errDone
				}
			case newErr := <-fileWatcher.Errors:
				// something wrong during the watching.
				logrus.Debug("unexpected error during watching file %v: %v", f.Name(), newErr)
				return err
			case <-watchTimeout.C:
				// FIXME: Since we hold the file handler in the process,
				// the fsnotify cannot propagate the Remove event.
				// This is workaround....
				//
				// More detail: https://github.com/fsnotify/fsnotify/issues/194
				_, sErr := os.Stat(f.Name())
				if sErr != nil {
					if os.IsNotExist(sErr) {
						return errDone
					}
					logrus.Debug("unexpected error during watching file %v: %v", f.Name(), sErr)
					return errDone
				}
			}
		}
	}

	// the dead loop to continue to read log
	for {
		msg, err := decodeOneLine()
		if err != nil {
			if err = handleError(err); err != nil {
				if err == errDone {
					return
				}

				watcher.Err <- err
				return
			}
			continue
		}

		if !cfg.Since.IsZero() && msg.Timestamp.Before(cfg.Since) {
			continue
		}

		if !cfg.Until.IsZero() && msg.Timestamp.After(cfg.Until) {
			return
		}

		select {
		case <-ctx.Done():
			return
		case watcher.Msgs <- msg:
		}
	}
}

// watchFileChange will watch the change of file.
func watchFileChange(filePath string) (*fsnotify.Watcher, error) {
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := fileWatcher.Add(filePath); err != nil {
		return nil, err
	}
	return fileWatcher, nil
}

// tailFile will read the log message until the io.EOF or limited by config.
func tailFile(r io.Reader, cfg *logger.ReadConfig, unmarshaler newUnmarshalFunc, watcher *logger.LogWatcher) {
	decodeOneLine := unmarshaler(r)

	for {
		msg, err := decodeOneLine()
		if err != nil {
			if err != io.EOF {
				watcher.Err <- err
			}
			return
		}

		if !cfg.Since.IsZero() && msg.Timestamp.Before(cfg.Since) {
			continue
		}

		if !cfg.Until.IsZero() && msg.Timestamp.After(cfg.Until) {
			return
		}

		select {
		case <-watcher.WatchClose():
			return
		case watcher.Msgs <- msg:
		}
	}
}

const (
	blockSize = 1024
	endOfLine = '\n'
)

// seekOffsetByTailLines is used to seek the offset in file by the number lines.
func seekOffsetByTailLines(rs io.ReadSeeker, n int) (int64, error) {
	if n <= 0 {
		return 0, nil
	}

	size, err := rs.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}

	var (
		block = -1
		cnt   = 0
		left  = int64(0)
		b     []byte

		readN int64
	)

	for {
		readN = int64(blockSize)
		left = size + int64(block*blockSize)
		if left < 0 {
			readN = int64(blockSize) + left
			left = 0
		}

		b = make([]byte, readN)
		if _, err := rs.Seek(left, os.SEEK_SET); err != nil {
			return 0, err
		}

		if _, err := rs.Read(b); err != nil {
			return 0, err
		}

		// if the line is enough or the file doesn't contain such lines
		cnt += bytes.Count(b, []byte{endOfLine})
		if cnt > n || left == 0 {
			break
		}
		block--
	}

	for cnt > n {
		if idx := bytes.IndexByte(b, endOfLine); idx >= 0 {
			left += int64(idx) + 1
			b = b[idx+1:]
		}
		cnt--
	}
	return left, nil
}
