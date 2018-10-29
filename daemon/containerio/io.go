package containerio

import (
	"io"
	"time"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/daemon/logger/crilog"
	"github.com/alibaba/pouch/pkg/multierror"
	"github.com/alibaba/pouch/pkg/streams"

	"github.com/containerd/containerd/cio"
	"github.com/sirupsen/logrus"
)

var (
	logcopierCloseTimeout = 10 * time.Second
	streamCloseTimeout    = 10 * time.Second
)

// wrapcio will wrap the DirectIO and IO.
//
// When the task exits, the containerd client will close the wrapcio.
type wrapcio struct {
	cio.IO

	ctrio *IO
}

func (wcio *wrapcio) Wait() {
	wcio.IO.Wait()
	wcio.ctrio.Wait()
}

func (wcio *wrapcio) Close() error {
	wcio.IO.Close()

	return wcio.ctrio.Close()
}

// IO represents the streams and logger.
type IO struct {
	id       string
	useStdin bool
	stream   *streams.Stream

	logdriver logger.LogDriver
	logcopier *logger.LogCopier
	criLog    *crilog.Log
}

// NewIO return IO instance.
func NewIO(id string, withStdin bool) *IO {
	s := streams.NewStream()
	if withStdin {
		s.NewStdinInput()
	} else {
		s.NewDiscardStdinInput()
	}

	return &IO{
		id:       id,
		useStdin: withStdin,
		stream:   s,
	}
}

// Reset reset the logdriver.
func (ctrio *IO) Reset() {
	if err := ctrio.Close(); err != nil {
		logrus.WithError(err).WithField("process", ctrio.id).
			Warnf("failed to close during reset IO")
	}

	if ctrio.useStdin {
		ctrio.stream.NewStdinInput()
	} else {
		ctrio.stream.NewDiscardStdinInput()
	}
	ctrio.logdriver = nil
	ctrio.logcopier = nil
	ctrio.criLog = nil
}

// SetLogDriver sets log driver to the IO.
func (ctrio *IO) SetLogDriver(logdriver logger.LogDriver) {
	ctrio.logdriver = logdriver
}

// Stream is used to export the stream field.
func (ctrio *IO) Stream() *streams.Stream {
	return ctrio.stream
}

// AttachCRILog will create CRILog and register it into stream.
func (ctrio *IO) AttachCRILog(path string, withTerminal bool) error {
	l, err := crilog.New(path, withTerminal)
	if err != nil {
		return err
	}

	// NOTE: it might write the same data into two different files, when
	// AttachCRILog is called for ReopenLog.
	ctrio.stream.AddStdoutWriter(l.Stdout)
	if l.Stderr != nil {
		ctrio.stream.AddStderrWriter(l.Stderr)
	}

	// NOTE: when close the previous crilog, it will evicted from the stream.
	if ctrio.criLog != nil {
		ctrio.criLog.Close()
	}
	ctrio.criLog = l
	return nil
}

// Wait wait for coping-data job.
func (ctrio *IO) Wait() {
	waitCh := make(chan struct{})
	go func() {
		defer close(waitCh)
		ctrio.stream.Wait()
	}()

	select {
	case <-waitCh:
	case <-time.After(streamCloseTimeout):
		logrus.Warnf("stream doesn't exit in time")
	}
}

// Close closes the stream and the logger.
func (ctrio *IO) Close() error {
	multiErrs := new(multierror.Multierrors)

	ctrio.Wait()
	if err := ctrio.stream.Close(); err != nil {
		multiErrs.Append(err)
	}

	if ctrio.logdriver != nil {
		if ctrio.logcopier != nil {
			waitCh := make(chan struct{})
			go func() {
				defer close(waitCh)
				ctrio.logcopier.Wait()
			}()
			select {
			case <-waitCh:
			case <-time.After(logcopierCloseTimeout):
				logrus.Warnf("logcopier doesn't exit in time")
			}
		}

		if err := ctrio.logdriver.Close(); err != nil {
			multiErrs.Append(err)
		}
	}

	if ctrio.criLog != nil {
		if err := ctrio.criLog.Close(); err != nil {
			multiErrs.Append(err)
		}
	}

	if multiErrs.Size() > 0 {
		return multiErrs
	}
	return nil
}

// InitContainerIO will start logger and coping data from fifo.
func (ctrio *IO) InitContainerIO(dio *DirectIO) (cio.IO, error) {
	if err := ctrio.startLogging(); err != nil {
		return nil, err
	}

	ctrio.stream.CopyPipes(streams.Pipes{
		Stdin:  dio.Stdin,
		Stdout: dio.Stdout,
		Stderr: dio.Stderr,
	})
	return &wrapcio{IO: dio, ctrio: ctrio}, nil
}

func (ctrio *IO) startLogging() error {
	if ctrio.logdriver == nil {
		return nil
	}

	ctrio.logcopier = logger.NewLogCopier(ctrio.logdriver, map[string]io.Reader{
		"stdout": ctrio.stream.NewStdoutPipe(),
		"stderr": ctrio.stream.NewStderrPipe(),
	})
	ctrio.logcopier.StartCopy()
	return nil
}
