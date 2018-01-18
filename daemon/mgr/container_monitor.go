package mgr

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	// EvExit represents container's exit event.
	EvExit = iota

	// TODO add more
)

// ContainerEvent represents the container's events.
type ContainerEvent struct {
	Kind   int
	c      *Container
	handle func(*Container) error
}

// String returns container's event type as a string.
func (e *ContainerEvent) String() string {
	switch e.Kind {
	case EvExit:
		return fmt.Sprintf("%s exit", e.c.ID())
	default:
		return "none"
	}
}

// WithHandle sets the event's handler.
func (e *ContainerEvent) WithHandle(handle func(*Container) error) *ContainerEvent {
	e.handle = handle
	return e
}

// ContainerExitEvent represents container's exit event.
func ContainerExitEvent(c *Container) *ContainerEvent {
	return &ContainerEvent{
		Kind: EvExit,
		c:    c,
	}
}

// TODO add more events

// ContainerMonitor is used to monitor contianer's event.
type ContainerMonitor struct {
	c chan *ContainerEvent
}

// NewContainerMonitor returns one ContainerMonitor object.
func NewContainerMonitor() *ContainerMonitor {
	m := &ContainerMonitor{
		c: make(chan *ContainerEvent, 1000),
	}

	go func() {
		for {
			ev := <-m.c

			logrus.Debugf("receive event: %s", ev)

			if ev.handle != nil {
				logrus.Infof("handle event: %s", ev)

				if err := ev.handle(ev.c); err != nil {
					logrus.Errorf("failed to handle event: %s", ev)
				}
			}
		}
	}()

	return m
}

// PostEvent sends a event to monitor.
func (m *ContainerMonitor) PostEvent(ev *ContainerEvent) {
	m.c <- ev
}
