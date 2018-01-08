package request

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/environment"
)

var defaultTimeout = time.Second * 10

// Option defines a type used to update http.Request.
type Option func(*http.Request) error

// WithHeader sets the Header of http.Request.
func WithHeader(key string, value string) Option {
	return func(r *http.Request) error {
		r.Header.Add(key, value)
		return nil
	}
}

// WithQuery sets the query field in URL.
func WithQuery(query url.Values) Option {
	return func(r *http.Request) error {
		r.URL.RawQuery = query.Encode()
		return nil
	}
}

// WithJSONBody encodes the input data to JSON and sets it to the body in http.Request
func WithJSONBody(obj interface{}) Option {
	return func(r *http.Request) error {
		b := bytes.NewBuffer([]byte{})

		if obj != nil {
			err := json.NewEncoder(b).Encode(obj)

			if err != nil {
				return err
			}
		}
		r.Body = ioutil.NopCloser(b)
		r.Header.Set("Content-Type", "application/json")
		return nil
	}
}

// DecodeBody decodes body to obj.
func DecodeBody(obj interface{}, body io.ReadCloser) error {
	// TODO: this fuction could only be called once
	defer body.Close()
	return json.NewDecoder(body).Decode(obj)
}

// Delete sends request to the default pouchd server with custom request options.
func Delete(endpoint string, opts ...Option) (*http.Response, error) {
	apiClient, err := newAPIClient(environment.PouchdAddress, environment.TLSConfig)
	if err != nil {
		return nil, err
	}

	req, err := newRequest(http.MethodDelete, apiClient.BaseURL()+endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return apiClient.HTTPCli.Do(req)
}

// Get sends request to the default pouchd server with custom request options.
func Get(endpoint string, opts ...Option) (*http.Response, error) {
	apiClient, err := newAPIClient(environment.PouchdAddress, environment.TLSConfig)
	if err != nil {
		return nil, err
	}

	req, err := newRequest(http.MethodGet, apiClient.BaseURL()+endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return apiClient.HTTPCli.Do(req)
}

// Post sends post request to pouchd.
func Post(endpoint string, opts ...Option) (*http.Response, error) {
	apiClient, err := newAPIClient(environment.PouchdAddress, environment.TLSConfig)
	if err != nil {
		return nil, err
	}

	req, err := newRequest(http.MethodPost, apiClient.BaseURL()+endpoint, opts...)
	if err != nil {
		return nil, err
	}

	// By default, if Content-Type in header is not set, set it to application/json
	if req.Header.Get("Content-Type") == "" {
		WithHeader("Content-Type", "application/json")(req)
	}
	return apiClient.HTTPCli.Do(req)
}

// newAPIClient return new HTTP client with tls.
//
// FIXME: Could we make some functions exported in alibaba/pouch/client?
func newAPIClient(host string, tls utils.TLSConfig) (*client.APIClient, error) {
	return client.NewAPIClient(host, tls)
}

// newRequest creates request targeting on specific host/path by method.
func newRequest(method, url string, opts ...Option) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		err := opt(req)
		if err != nil {
			return nil, err
		}
	}
	//fmt.Println(req)

	return req, nil
}

// Hijack posts hijack request.
func Hijack(endpoint string, opts ...Option) (*http.Response, net.Conn, *bufio.Reader, error) {
	req, err := newRequest(http.MethodPost, endpoint, opts...)
	if err != nil {
		return nil, nil, nil, err
	}

	req.Host = environment.PouchdUnixDomainSock
	conn, err := net.DialTimeout("unix", req.Host, defaultTimeout)
	if err != nil {
		return nil, nil, nil, err
	}

	clientconn := httputil.NewClientConn(conn, nil)
	defer clientconn.Close()

	// For hijack request, pouchd server will return 200 status or
	// 101 status for switching protocol. For the 200 status code, http
	// proto has definied:
	//
	//	If there is no Content-Length or chunked Transfer-Encoding on
	//	a *Response and the status is not 1xx, 204 or 304, then the
	//	body is unbounded.
	//
	//	For this case, the response is always terminated by the first
	//	empty line after the header fields.
	//	More details in RFC 2616, section 4.4.
	//
	// "persistent connection closed" is expectd error for 200 status.
	resp, err := clientconn.Do(req)
	if err != httputil.ErrPersistEOF {
		if err != nil {
			return nil, nil, nil, err
		}

		if resp.StatusCode != http.StatusSwitchingProtocols {
			resp.Body.Close()
			return nil, nil, nil, fmt.Errorf("unable to upgrade proto, got http status: %v", resp.StatusCode)
		}
	}

	rwc, br := clientconn.Hijack()
	return resp, rwc, br, nil
}
