package errtypes

import (
	"github.com/pkg/errors"
)

var (
	// ErrInvalidParam represents the parameters are invalid.
	ErrInvalidParam = errorType{codeInvalidParam, "invalid param"}

	// ErrNotfound represents the object is not found, not exist.
	ErrNotfound = errorType{codeNotFound, "not found"}

	// ErrAlreadyExisted represents the object has already existed.
	ErrAlreadyExisted = errorType{codeAlreadyExisted, "already existed"}

	// ErrConflict represents the parameters are invalid.
	ErrConflict = errorType{codeConflict, "conflict"}

	// ErrTooMany reprensents the objects are too many.
	ErrTooMany = errorType{codeTooMany, "too many"}

	// ErrTimeout represents the operation is time out.
	ErrTimeout = errorType{codeTimeout, "time out"}

	// ErrLockfailed represents that failed to lock.
	ErrLockfailed = errorType{codeLockfailed, "lock failed"}

	// ErrNotImplemented represents that the function is not implemented.
	ErrNotImplemented = errorType{codeNotImplemented, "not implemented"}

	// ErrInUse represents that object is using.
	ErrInUse = errorType{codeInUse, "in use"}

	// ErrNotModified represents that the resource is not modified
	ErrNotModified = errorType{codeNotModified, "not modified"}

	// ErrPreCheckFailed represents that failed to pre check.
	ErrPreCheckFailed = errorType{codePreCheckFailed, "pre check failed"}
)

const (
	codeInvalidParam = iota
	codeNotFound
	codeAlreadyExisted
	codeConflict
	codeTooMany
	codeTimeout
	codeLockfailed
	codeNotImplemented
	codeInUse
	codeNotModified
	codePreCheckFailed

	// volume error code
	codeVolumeExisted
	codeVolumeDriverNotFound
	codeVolumeMetaNotFound
)

type errorType struct {
	code int
	err  string
}

func (e errorType) Error() string {
	return e.err
}

// IsNotfound checks the error is object Notfound or not.
func IsNotfound(err error) bool {
	return checkError(err, codeNotFound)
}

// IsAlreadyExisted checks the error is object AlreadyExisted or not.
func IsAlreadyExisted(err error) bool {
	return checkError(err, codeAlreadyExisted)
}

// IsInvalidParam checks the error is the parameters are invalid or not.
func IsInvalidParam(err error) bool {
	return checkError(err, codeInvalidParam)
}

// IsTimeout checks the error is time out or not.
func IsTimeout(err error) bool {
	return checkError(err, codeTimeout)
}

// IsInUse checks the error is using by others or not.
func IsInUse(err error) bool {
	return checkError(err, codeInUse)
}

// IsNotModified checks the error is not modified error or not.
func IsNotModified(err error) bool {
	return checkError(err, codeNotModified)
}

// IsPreCheckFailed checks the error is failed to pre check or not.
func IsPreCheckFailed(err error) bool {
	return checkError(err, codePreCheckFailed)
}

func checkError(err error, code int) bool {
	err = causeError(err)

	if err0, ok := err.(errorType); ok && err0.code == code {
		return true
	}
	return false
}

func causeError(err error) error {
	return errors.Cause(err)
}
