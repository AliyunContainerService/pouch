package ctrd

import (
	"context"
	"sync"
	"time"

	"github.com/alibaba/pouch/pkg/errtypes"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Message is used to watch containerd.
type Message struct {
	exitCode uint32
	exitTime time.Time
	err      error
}

// RawError returns the error contained in Message.
func (m *Message) RawError() error {
	return m.err
}

// ExitCode returns the exit code in Message.
func (m *Message) ExitCode() uint32 {
	return m.exitCode
}

// ExitTime returns the exit time in Message.
func (m *Message) ExitTime() time.Time {
	return m.exitTime
}

type watch struct {
	sync.Mutex
	containers map[string]*containerPack
	hooks      []func(string, *Message) error

	// containerdDead to specify whether containerd process is dead
	containerdDead bool
}

func (w *watch) setContainerdDead(isDead bool) error {
	w.Lock()
	defer w.Unlock()

	w.containerdDead = isDead
	return nil
}

func (w *watch) isContainerdDead() bool {
	w.Lock()
	defer w.Unlock()

	return w.containerdDead
}

func (w *watch) add(pack *containerPack) {
	w.Lock()
	defer w.Unlock()

	// TODO(ziren): AcquireQuota may occurred an error
	// record stream client for grpc client.
	_ = pack.client.Consume(1)

	w.containers[pack.id] = pack

	go func(w *watch, pack *containerPack) {
		status := <-pack.sch

		// if containerd is dead, the task.Wait channel return is not because
		// container' task quit, but the channel has broken.
		// so we just return.
		if w.isContainerdDead() {
			return
		}

		// isContainerdDead only take effect when contained stop normal, if containerd
		// stop unexpected, judge exit time is zero, zero exit time means grpc connection
		// is broken.
		if status.ExitTime().IsZero() {
			return
		}

		logrus.Infof("the task has quit, id: %s, err: %v, exitcode: %d, time: %v",
			pack.id, status.Error(), status.ExitCode(), status.ExitTime())

		// Also should release quota when the container destroyed
		// We should release quota of client that the pack is in using,
		// not the grpc client executing this parts of code.
		pack.client.Produce(1)

		msg := &Message{
			err:      status.Error(),
			exitCode: status.ExitCode(),
			exitTime: status.ExitTime(),
		}

		if !pack.skipStopHooks {
			for _, hook := range w.hooks {
				if err := hook(pack.id, msg); err != nil {
					logrus.Errorf("failed to execute the exit hooks: %v", err)
					break
				}
			}
		}

		// NOTE: we should delete task/container after update the status, for example, status code.
		if _, err := pack.task.Delete(context.Background()); err != nil {
			logrus.Errorf("failed to delete task, container id: %s: %v", pack.id, err)
		}

		if err := pack.container.Delete(context.Background()); err != nil {
			logrus.Errorf("failed to delete container, container id: %s: %v", pack.id, err)
		}
		pack.ch <- msg

	}(w, pack)

	logrus.Infof("success to add container, id: %s", pack.id)
}

func (w *watch) remove(ctx context.Context, id string) error {
	w.Lock()
	defer w.Unlock()

	delete(w.containers, id)
	return nil
}

func (w *watch) get(id string) (*containerPack, error) {
	w.Lock()
	defer w.Unlock()

	pack, ok := w.containers[id]
	if !ok {
		return pack, errors.Wrapf(errtypes.ErrNotfound, "container %s in metadata", id)
	}
	return pack, nil
}

func (w *watch) notify(id string) chan *Message {
	w.Lock()
	defer w.Unlock()

	pack, ok := w.containers[id]
	if !ok {
		ch := make(chan *Message, 1)
		ch <- &Message{
			err: errors.Wrapf(errtypes.ErrNotfound, "container %s in metadata", id),
		}
		return ch
	}
	return pack.ch
}
