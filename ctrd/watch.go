package ctrd

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/log"

	"github.com/containerd/containerd"
	"github.com/pkg/errors"
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
	hooks      []func(string, *Message, func() error) error

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

// isChannelClosed means containerd break unexpected, all exit channel
// ctrd watched will exit.
func isChannelClosed(s containerd.ExitStatus) bool {
	return s.ExitTime().IsZero() && strings.Contains(s.Error().Error(), "transport is closing")
}

func (w *watch) add(ctx context.Context, pack *containerPack) {
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
		// stop unexpected, judge whether channel is broken, if does, skip this message,
		// since task still running
		if isChannelClosed(status) {
			log.With(ctx).Warnf("receive exit message since channel broken, %+v", status)
			return
		}

		log.With(ctx).Infof("the task has quit, id: %s, err: %v, exitcode: %d, time: %v",
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

		// NOTE: cleanup action should be taken only once!
		var cleanupOnce sync.Once
		cleanupFunc := func() error {
			cleanupOnce.Do(func() {
				if _, err := pack.task.Delete(context.Background()); err != nil {
					log.With(ctx).Errorf("failed to delete task, container id: %s: %v", pack.id, err)
				}

				if err := pack.container.Delete(context.Background()); err != nil {
					log.With(ctx).Errorf("failed to delete container, container id: %s: %v", pack.id, err)
				}
			})
			return nil
		}

		pack.l.RLock()
		skipCleanup := pack.skipStopHooks
		pack.l.RUnlock()
		if !skipCleanup {
			for _, hook := range w.hooks {
				if err := hook(pack.id, msg, cleanupFunc); err != nil {
					log.With(ctx).Errorf("failed to execute the exit hooks: %v", err)
					break
				}
			}

			// if stop container was triggered, skipStopHooks will be set to true, cleanup logic will be invoke by stop
			// routine.
			cleanupFunc()
		}

		pack.ch <- msg

	}(w, pack)

	log.With(ctx).Infof("success to add container")
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
