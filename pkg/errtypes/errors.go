package errtypes

import (
	"github.com/pkg/errors"
)

var (
	// ErrNotfound represents the object is not found, not exist.
	ErrNotfound = errorType{codeNotfound, "not found"}

	// ErrAlreadyExisted represents the object has already existed.
	ErrAlreadyExisted = errorType{codeAlreadyExisted, "already existed"}

	// ErrInvalidParam represents the parameters are invalid.
	ErrInvalidParam = errorType{codeInvalidParam, "invalid param"}

	// ErrTooMany reprensents the objects are too many.
	ErrTooMany = errorType{codeTooMany, "too many"}

	// ErrInvalidType represents the object's type is invalid.
	ErrInvalidType = errorType{codeInvalidType, "invalid type"}

	// ErrTimeout represents the operation is time out.
	ErrTimeout = errorType{codeTimeout, "time out"}

	// ErrLockfailed represents that failed to lock.
	ErrLockfailed = errorType{codeLockfailed, "lock failed"}

	// ErrNotImplemented represents that the function is not implemented.
	ErrNotImplemented = errorType{codeNotImplemented, "not implemented"}

	// ErrUsingbyContainers represents that object is used by containers
	ErrUsingbyContainers = errorType{codeInUse, "using by containers"}
)

const (
	codeNotfound = iota
	codeAlreadyExisted
	codeInvalidParam
	codeTooMany
	codeInvalidType
	codeTimeout
	codeLockfailed
	codeNotImplemented
	codeInUse
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
	return checkError(err, codeNotfound)
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
