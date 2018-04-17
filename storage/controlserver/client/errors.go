package client

import (
	"net/http"
)

// Error is defined as client error.
type Error struct {
	status int
	err    string
}

func newError(status int, err string) Error {
	return Error{
		status: status,
		err:    err,
	}
}

// Error returns client Error's error message.
func (e Error) Error() string {
	return e.err
}

// IsNotFound will check the error is "StatusNotFound" or not.
func (e Error) IsNotFound() bool {
	return e.status == http.StatusNotFound
}
