package containerio

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/daemon/logger/jsonfile"

	"github.com/sirupsen/logrus"
)

func init() {
	Register(func() Backend {
		return &jsonFile{}
	})
}

var jsonFilePathName = "json.log"

// TODO(fuwei): add compress/logrotate configure
type jsonFile struct {
	closed bool

	copier *jsonfile.JSONLogFile

	stdoutWriter io.WriteCloser
	stderrWriter io.WriteCloser
}

func (jf *jsonFile) Name() string {
	return "jsonfile"
}

func (jf *jsonFile) Init(opt *Option) error {
	if _, err := os.Stat(opt.rootDir); err != nil {
		return err
	}

	logPath := filepath.Join(opt.rootDir, jsonFilePathName)
	w, err := jsonfile.NewJSONLogFile(logPath, 0644)
	if err != nil {
		return err
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	jf.copier, jf.stdoutWriter, jf.stderrWriter = w, stdoutWriter, stderrWriter
	go jf.copy("stdout", stdoutReader)
	go jf.copy("stderr", stderrReader)

	return nil
}

func (jf *jsonFile) In() io.Reader {
	return nil
}

func (jf *jsonFile) Out() io.Writer {
	return jf.stdoutWriter
}

func (jf *jsonFile) Err() io.Writer {
	return jf.stderrWriter
}

func (jf *jsonFile) Close() error {
	if jf.closed {
		return nil
	}

	if err := jf.stdoutWriter.Close(); err != nil {
		return err
	}

	if err := jf.stderrWriter.Close(); err != nil {
		return err
	}

	if err := jf.copier.Close(); err != nil {
		return err
	}

	jf.closed = true
	return nil
}

func (jf *jsonFile) copy(source string, reader io.ReadCloser) {
	var (
		bs  []byte
		err error

		firstPartial = true
		isPartial    bool
		createdTime  time.Time

		defaultBufSize = 16 * 1024
	)

	defer reader.Close()
	br := bufio.NewReaderSize(reader, defaultBufSize)

	for {
		bs, isPartial, err = br.ReadLine()
		if err != nil {
			if err != io.EOF {
				logrus.Errorf("failed to copy %v message into jsonfile: %v", source, err)
			}
			return
		}

		// NOTE: The partial content will share the same timestamp.
		if firstPartial {
			createdTime = time.Now().UTC()
		}

		if isPartial {
			firstPartial = false
		} else {
			firstPartial = true
			bs = append(bs, '\n')
		}

		if err = jf.copier.WriteLogMessage(&logger.LogMessage{
			Source:    source,
			Line:      bs,
			Timestamp: createdTime,
		}); err != nil {
			logrus.Errorf("failed to copy %v message into jsonfile: %v", source, err)
			return
		}
	}
}
