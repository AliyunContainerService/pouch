package main

import (
	"net/http"
)

// Request wrap the http.Rquest and some other states.
type Request struct {
	req *http.Request
	cli *http.Client
}

// Send send a http request to server, any errors will be wrapped into Response.
func (r *Request) Send() *Response {
	resp, err := r.cli.Do(r.req)
	if err != nil {
		return &Response{
			Err: err,
		}
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}

	if resp.StatusCode == http.StatusNotFound {
		response.Err = ErrHTTPNotfound
		return response
	}
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		response.Err = ErrHTTP4xx
		return response
	}
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		response.Err = ErrHTTP5xx
		return response
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		response.Err = ErrHTTPNotOK
		return response
	}

	response.Body = resp.Body
	return response
}
