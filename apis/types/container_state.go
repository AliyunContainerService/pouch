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

// String returns the container's status by string.
func (s ContainerStatus) String() string {
	switch s {
	case CREATED:
		return "Created"
	case RUNNING:
		return "Running"
	case PAUSED:
		return "Paused"
	case RESTARTING:
		return "Restarting"
	case OOMKILLED:
		return "OOMKilled"
	case REMOVALINPROGRESS:
		return "Removing"
	case STOPPING:
		return "Stopping"
	case STOPPED:
		return "Stopped"
	case DEAD:
		return "Dead"
	default:
		return "Unknown"
	}
}
