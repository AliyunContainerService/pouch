package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

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

// WithJSONBody sets the body in http.Request
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
		return nil
	}
}

// DecodeToStruct decodes body to obj.
func DecodeToStruct(c *check.C, body io.ReadCloser, obj interface{}) {
	err := json.NewDecoder(body).Decode(obj)
	c.Assert(err, check.IsNil)
}

// Delete sends request to the default pouchd server with custom request options.
func Delete(c *check.C, endpoint string, opts ...Option) (*http.Response, error) {
	return prepareClientAndInitRequest(c, http.MethodDelete, endpoint, opts...)
}

// Get sends request to the default pouchd server with custom request options.
func Get(c *check.C, endpoint string, opts ...Option) (*http.Response, error) {
	return prepareClientAndInitRequest(c, http.MethodGet, endpoint, opts...)
}

// Post sends post request to pouchd.
func Post(c *check.C, endpoint string, opts ...Option) (*http.Response, error) {
	return prepareClientAndInitRequest(c, http.MethodPost, endpoint, opts...)
}

func prepareClientAndInitRequest(c *check.C, method string, path string, opts ...Option) (*http.Response, error) {
	apiClient, err := newAPIClient(environment.PouchdAddress, environment.TLSConfig)
	c.Assert(err, check.IsNil)

	req, err := newRequest(method, apiClient.BaseURL()+path, opts...)
	c.Assert(err, check.IsNil)

	resp, err := apiClient.HTTPCli.Do(req)
	c.Assert(err, check.IsNil)

	// if status code is bigger than 400, error message is in response body
	if resp.StatusCode >= 400 {
		errMsg := types.Error{}
		err := json.NewDecoder(resp.Body).Decode(&errMsg)
		c.Assert(err, check.IsNil)
		return resp, fmt.Errorf(errMsg.Message)
	}

	return resp, fmt.Errorf("no error")
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
	return req, nil
}
