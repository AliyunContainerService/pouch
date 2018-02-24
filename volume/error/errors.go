package error

const (
	volumeNotFound errCode = iota
	volumeExisted
	storageNotFound
	driverNotFound
	localMetaNotFound
	disableControl
)

var (
	// ErrVolumeNotFound represents error is "volume not found"
	ErrVolumeNotFound = CoreError{volumeNotFound, "volume not found"}

	// ErrVolumeExisted represents error is "volume exist"
	ErrVolumeExisted = CoreError{volumeExisted, "volume exist"}

	// ErrStorageNotFound represents error is "storage not found"
	ErrStorageNotFound = CoreError{storageNotFound, "storage not found"}

	// ErrDriverNotFound represents error is "driver not found"
	ErrDriverNotFound = CoreError{driverNotFound, "driver not found"}

	// ErrLocalMetaNotFound represents error is "local meta not found"
	ErrLocalMetaNotFound = CoreError{localMetaNotFound, "local meta not found"}

	// ErrDisableControl represents error is "disable control server"
	ErrDisableControl = CoreError{disableControl, "disable control server"}
)

type errCode int

// CoreError represents volume core error struct.
type CoreError struct {
	ec  errCode
	err string
}

// Error returns core error message.
func (e CoreError) Error() string {
	return e.err
}

// IsVolumeNotFound is used to check error is volumeNotFound or not.
func (e CoreError) IsVolumeNotFound() bool {
	return e.ec == volumeNotFound
}

// IsStorageNotFound is used to check error is storageNotFound or not.
func (e CoreError) IsStorageNotFound() bool {
	return e.ec == storageNotFound
}

// IsDriverNotFound is used to check error is driverNotFound or not.
func (e CoreError) IsDriverNotFound() bool {
	return e.ec == driverNotFound
}

// IsVolumeExisted is used to check error is volumeExisted or not.
func (e CoreError) IsVolumeExisted() bool {
	return e.ec == volumeExisted
}

// IsLocalMetaNotFound is used to check error is localMetaNotFound or not.
func (e CoreError) IsLocalMetaNotFound() bool {
	return e.ec == localMetaNotFound
}

// IsDisableControl is used to check error is disableControl or not.
func (e CoreError) IsDisableControl() bool {
	return e.ec == disableControl
}
