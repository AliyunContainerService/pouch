package client

import (
	"bufio"
	"bytes"
	"context"
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

func (client *APIClient) get(ctx context.Context, path string, query url.Values, headers map[string][]string) (*Response, error) {
	return client.sendRequest(ctx, "GET", path, query, nil, headers)
}

func (client *APIClient) post(ctx context.Context, path string, query url.Values, obj interface{}, headers map[string][]string) (*Response, error) {
	return client.sendRequest(ctx, "POST", path, query, obj, headers)
}

func (client *APIClient) delete(ctx context.Context, path string, query url.Values, headers map[string][]string) (*Response, error) {
	return client.sendRequest(ctx, "DELETE", path, query, nil, headers)
}

func (client *APIClient) hijack(ctx context.Context, path string, query url.Values, obj interface{}, header map[string][]string) (net.Conn, *bufio.Reader, error) {
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

	fullPath := client.baseURL + client.GetAPIPath(path, query)
	req, err := http.NewRequest(method, fullPath, body)
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

func (client *APIClient) sendRequest(ctx context.Context, method, path string, query url.Values, obj interface{}, headers map[string][]string) (*Response, error) {
	req, err := client.newRequest(method, path, query, obj, headers)
	if err != nil {
		return nil, err
	}

	resp, err := cancellableDo(ctx, client.HTTPCli, req)
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

func cancellableDo(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	type contextResp struct {
		response *http.Response
		err      error
	}

	ctxResp := make(chan contextResp, 1)
	go func() {
		resp, err := client.Do(req)
		ctxResp <- contextResp{
			response: resp,
			err:      err,
		}
	}()

	select {
	case <-ctx.Done():
		tr := client.Transport.(*http.Transport)
		tr.CancelRequest(req)
		<-ctxResp
		return nil, ctx.Err()

	case resp := <-ctxResp:
		return resp.response, resp.err
	}
}
