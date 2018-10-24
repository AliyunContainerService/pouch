package errtypes

var (
	// ErrVolumeInUse represents that volume in use.
	ErrVolumeInUse = errorType{codeInUse, "volume is in use"}

	// ErrVolumeNotFound represents that no such volume.
	ErrVolumeNotFound = errorType{codeNotFound, "no such volume"}

	// ErrVolumeExisted represents error is "volume exist"
	ErrVolumeExisted = errorType{codeVolumeExisted, "volume exist"}

	// ErrVolumeDriverNotFound represents error is "driver not found"
	ErrVolumeDriverNotFound = errorType{codeVolumeDriverNotFound, "driver not found"}

	// ErrVolumeMetaNotFound represents error is "local meta not found"
	ErrVolumeMetaNotFound = errorType{codeVolumeMetaNotFound, "local meta not found"}
)

// IsVolumeInUse is used to check error is volume in use.
func IsVolumeInUse(err error) bool {
	return checkError(err, codeInUse)
}

// IsVolumeNotFound is used to check error is volumeNotFound or not.
func IsVolumeNotFound(err error) bool {
	return checkError(err, codeNotFound)
}

// IsVolumeExisted is used to check error is volumeExisted or not.
func IsVolumeExisted(err error) bool {
	return checkError(err, codeVolumeExisted)
}

// IsVolumeDriverNotFound is used to check error is driverNotFound or not.
func IsVolumeDriverNotFound(err error) bool {
	return checkError(err, codeVolumeDriverNotFound)
}

// IsVolumeMetaNotFound is used to check error is localMetaNotFound or not.
func IsVolumeMetaNotFound(err error) bool {
	return checkError(err, codeVolumeMetaNotFound)
}
