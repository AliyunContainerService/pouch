package error

const (
	volumeNotfound errCode = iota
	volumeExisted
	storageNotfound
	driverNotfound
	localMetaNotfound
	disableControl
)

var (
	// ErrVolumeNotfound represents error is "volume not found"
	ErrVolumeNotfound = CoreError{volumeNotfound, "volume not found"}

	// ErrVolumeExisted represents error is "volume exist"
	ErrVolumeExisted = CoreError{volumeExisted, "volume exist"}

	// ErrStorageNotfound represents error is "storage not found"
	ErrStorageNotfound = CoreError{storageNotfound, "storage not found"}

	// ErrDriverNotfound represents error is "driver not found"
	ErrDriverNotfound = CoreError{driverNotfound, "driver not found"}

	// ErrLocalMetaNotfound represents error is "local meta not found"
	ErrLocalMetaNotfound = CoreError{localMetaNotfound, "local meta not found"}

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

// IsVolumeNotfound is used to check error is volumeNotfound or not.
func (e CoreError) IsVolumeNotfound() bool {
	return e.ec == volumeNotfound
}

// IsStorageNotfound is used to check error is storageNotfound or not.
func (e CoreError) IsStorageNotfound() bool {
	return e.ec == storageNotfound
}

// IsDriverNotfound is used to check error is driverNotfound or not.
func (e CoreError) IsDriverNotfound() bool {
	return e.ec == driverNotfound
}

// IsVolumeExisted is used to check error is volumeExisted or not.
func (e CoreError) IsVolumeExisted() bool {
	return e.ec == volumeExisted
}

// IsLocalMetaNotfound is used to check error is localMetaNotfound or not.
func (e CoreError) IsLocalMetaNotfound() bool {
	return e.ec == localMetaNotfound
}

// IsDisableControl is used to check error is disableControl or not.
func (e CoreError) IsDisableControl() bool {
	return e.ec == disableControl
}
