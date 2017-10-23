package ctrd

const (
	containerNotfound = iota
	taskNotfound
	imageNotfound
	trylockFailed
	timeoutFailed
)

var (
	// ErrContainerNotfound is returned when container is not found from containerd.
	ErrContainerNotfound = Error{containerNotfound, "Container not found"}
	// ErrTaskNotfound is returned when task is not found from containerd.
	ErrTaskNotfound = Error{taskNotfound, "Task not found"}
	// ErrImageNotfound is returned when image is not found from containerd.
	ErrImageNotfound = Error{imageNotfound, "Image not found"}
	// ErrTrylockFailed is returned when trylock failed.
	ErrTrylockFailed = Error{trylockFailed, "Trylock failed"}
	// ErrTimeout is returned when the connection time out.
	ErrTimeout = Error{timeoutFailed, "time out"}
)

// Error contains a code and a message. It implements the error interface.
type Error struct {
	code int
	msg  string
}

// Error returns the error message.
func (e Error) Error() string {
	return e.msg
}

// IsNotfound return true if this is exactly something not found.
func (e Error) IsNotfound() bool {
	switch e.code {
	case containerNotfound:
		return true
	case taskNotfound:
		return true
	case imageNotfound:
		return true
	}
	return false
}

// IsTrylockFailed determines whether the error code is trylockFailed.
func (e Error) IsTrylockFailed() bool {
	return e.code == trylockFailed
}

// IsTimeout determines whether the error code is timeoutFailed.
func (e Error) IsTimeout() bool {
	return e.code == timeoutFailed
}
