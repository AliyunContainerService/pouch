package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
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

func (client *APIClient) get(path string, query url.Values) (*Response, error) {
	return client.sendRequest("GET", path, query, nil)
}

func (client *APIClient) post(path string, query url.Values, obj interface{}) (*Response, error) {
	return client.sendRequest("POST", path, query, obj)
}

func (client *APIClient) delete(path string, query url.Values) (*Response, error) {
	return client.sendRequest("DELETE", path, query, nil)
}

func (client *APIClient) hijack(path string, query url.Values, obj interface{}, header map[string][]string) (net.Conn, *bufio.Reader, error) {
	req, err := client.newRequest("POST", path, query, obj, header)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")

	req.Host = client.addr
	conn, err := net.DialTimeout(client.proto, client.addr, defaultTimeout)
	if err != nil {
		return nil, nil, err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	clientconn := httputil.NewClientConn(conn, nil)
	defer clientconn.Close()

	if _, err := clientconn.Do(req); err != nil {
		return nil, nil, err
	}

	rwc, br := clientconn.Hijack()

	return rwc, br, nil
}

func (client *APIClient) newRequest(method, path string, query url.Values, obj interface{}, header map[string][]string) (*http.Request, error) {
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

	req, err := http.NewRequest(method, client.baseURL+getAPIPath(path, query), body)
	if err != nil {
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			req.Header[k] = v
		}
	}

	return req, err
}

func (client *APIClient) sendRequest(method, path string, query url.Values, obj interface{}) (*Response, error) {
	req, err := client.newRequest(method, path, query, obj, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.HTTPCli.Do(req)
	if err != nil {
		return nil, err
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
