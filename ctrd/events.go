package ctrd

import (
	"github.com/containerd/containerd/runtime"
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
