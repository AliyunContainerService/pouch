package main

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

var (
	// ErrHTTPNotfound represents the 404 error of a http request.
	ErrHTTPNotfound = RespError{codeHTTPNotfound, "404: not found"}
	// ErrHTTP4xx represents 4xx errors of a http request.
	ErrHTTP4xx = RespError{codeHTTP4xx, "4xx: request error"}
	// ErrHTTP5xx represents 5xx errors of a http request.
	ErrHTTP5xx = RespError{codeHTTP5xx, "5xx: server error"}
	// ErrHTTPNotOK represents other errors(not 200) of a http request.
	ErrHTTPNotOK = RespError{codeHTTPNotOK, "not ok"}
)

const (
	codeHTTPNotfound = iota
	codeHTTP4xx
	codeHTTP5xx
	codeHTTPNotOK
)

// RespError define the response error.
type RespError struct {
	code int
	msg  string
}

// Error implement the error interface.
func (e RespError) Error() string {
	return e.msg
}

// Response wrap the http.Response and other states.
type Response struct {
	Err        error
	StatusCode int
	Status     string
	Body       io.ReadCloser
}

// Error return error state.
func (r *Response) Error() error {
	return r.Err
}

// DecodeBody receive all body data and decode them with json.
func (r *Response) DecodeBody(obj interface{}, dec ...func(interface{}, io.Reader) error) error {
	deserialize := func(o interface{}, rd io.Reader) error {
		return json.NewDecoder(rd).Decode(o)
	}
	if len(dec) != 0 {
		deserialize = dec[0]
	}

	if err := deserialize(obj, r.Body); err != nil {
		if err == io.EOF {
			return io.EOF
		}
		return errors.Wrap(err, "failed to decode body")
	}
	return nil
}

// Close close the 'Body' io.
func (r *Response) Close() error {
	return r.Body.Close()
}
