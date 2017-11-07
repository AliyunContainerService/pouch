package client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	// ErrHTTPNotfound represents the 404 error of a http request.
	ErrHTTPNotfound = RespError{codeHTTPNotfound, "404: not found"}
)

const (
	codeHTTPNotfound = iota
)

// RespError defines the response error.
type RespError struct {
	code int
	msg  string
}

// Error implements the error interface.
func (e RespError) Error() string {
	return e.msg
}

// Response wraps the http.Response and other states.
type Response struct {
	StatusCode int
	Status     string
	Body       io.ReadCloser
}

func (cli *Client) get(path string, query url.Values) (*Response, error) {
	return cli.sendRequest("GET", path, query, nil)
}

func (cli *Client) post(path string, query url.Values, obj interface{}) (*Response, error) {
	return cli.sendRequest("POST", path, query, obj)
}

func (cli *Client) delete(path string, query url.Values) (*Response, error) {
	return cli.sendRequest("DELETE", path, query, nil)
}

func (cli *Client) sendRequest(method, path string, query url.Values, obj interface{}) (*Response, error) {
	var body io.Reader
	if method == "POST" {
		if obj != nil {
			b, err := json.Marshal(obj)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(b)
		} else {
			body = bytes.NewReader([]byte{})
		}
	}

	req, err := http.NewRequest(method, cli.baseURL+getAPIPath(path, query), body)
	if err != nil {
		return nil, err
	}

	resp, err := cli.httpCli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrHTTPNotfound
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, RespError{code: resp.StatusCode, msg: string(data)}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       resp.Body,
	}, nil
}

func getAPIPath(path string, query url.Values) string {
	u := url.URL{
		Path: path,
	}

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	return u.String()
}
