package ctrd

import (
	"context"

	eventsapi "github.com/containerd/containerd/api/services/events/v1"
	"github.com/containerd/containerd/runtime"
	"github.com/pkg/errors"
)

const (
	// ContainersCreateEventTopic for container create
	ContainersCreateEventTopic = "/containers/create"
	// ContainersDeleteEventTopic for container delete
	ContainersDeleteEventTopic = "/containers/delete"

	// TaskCreateEventTopic for task create
	TaskCreateEventTopic = runtime.TaskCreateEventTopic
	// TaskDeleteEventTopic for task delete
	TaskDeleteEventTopic = runtime.TaskDeleteEventTopic
	// TaskExitEventTopic for task exit
	TaskExitEventTopic = runtime.TaskExitEventTopic
	// TaskOOMEventTopic for task oom
	TaskOOMEventTopic = runtime.TaskOOMEventTopic
)

// Events subscribe containerd events through an event subscribe client.
func (c *Client) Events(ctx context.Context, ef ...string) (eventsapi.Events_SubscribeClient, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, ErrGetCtrdClient.Error())
	}

	eventsClient := wrapperCli.client.EventService()
	return eventsClient.Subscribe(ctx, &eventsapi.SubscribeRequest{
		Filters: ef,
	})
}
