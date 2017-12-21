package request

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/environment"
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
