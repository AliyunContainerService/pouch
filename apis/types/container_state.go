package types

import "time"

// ContainerStatus is a enum type and indicates container status
type ContainerStatus int

/* container status enums */
const (
	CREATED ContainerStatus = iota
	RUNNING
	PAUSED
	RESTARTING
	OOMKILLED
	REMOVALINPROGRESS
	STOPPING
	STOPPED
	DEAD
)

// ContainerState houses container status, start and stop information
type ContainerState struct {
	StartedAt     time.Time
	Status        ContainerStatus
	FinishedAt    time.Time
	Pid           int
	ExitCodeValue int    `json:"ExitCode"`
	ErrorMsg      string `json:"Error"` // contains last known error when starting the container
}
