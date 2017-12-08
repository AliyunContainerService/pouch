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
)

const (
	codeNotfound = iota
	codeAlreadyExisted
	codeInvalidParam
	codeTooMany
	codeInvalidType
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
	err = causeError(err)

	if err0, ok := err.(errorType); ok && err0.code == codeNotfound {
		return true
	}
	return false
}

// IsAlreadyExisted checks the error is object AlreadyExisted or not.
func IsAlreadyExisted(err error) bool {
	err = causeError(err)

	if err0, ok := err.(errorType); ok && err0.code == codeAlreadyExisted {
		return true
	}
	return false
}

// IsInvalidParam checks the error is the parameters are invalid or not.
func IsInvalidParam(err error) bool {
	err = causeError(err)

	if err0, ok := err.(errorType); ok && err0.code == codeInvalidParam {
		return true
	}
	return false
}

func causeError(err error) error {
	return errors.Cause(err)
}
